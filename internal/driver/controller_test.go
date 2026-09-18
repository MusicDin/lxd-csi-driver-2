package driver

import (
	"context"
	"maps"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/canonical/lxd-csi-driver/internal/devlxd"
	lxdClient "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

func TestCreateVolumeMultiNodeAccessMode(t *testing.T) {
	tests := []struct {
		Name          string
		StorageDriver string
		expectError   string
	}{
		{
			Name:          "Ensure multi node multi writer volume is created on cephfs driver",
			StorageDriver: "cephfs",
			expectError:   "",
		},
		{
			Name:          "Ensure multi node multi writer volume is rejected on local driver",
			StorageDriver: "dir",
			expectError:   `Access mode "MULTI_NODE_MULTI_WRITER" is not supported by storage driver "dir"`,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			d := &Driver{
				name:     "lxd.csi.canonical.com",
				version:  "test",
				endpoint: "unix:///csi/csi.sock",
				nodeID:   "test-node",
			}

			var calledCreate bool
			d.devLXD = &devlxd.FakeServer{
				GetPoolFunc: func(pool string) (*api.DevLXDStoragePool, string, error) {
					return &api.DevLXDStoragePool{Name: pool, Driver: test.StorageDriver}, "", nil
				},
				GetStateFunc: func() (*api.DevLXDGet, error) {
					return &api.DevLXDGet{
						DevLXDGetUntrusted: api.DevLXDGetUntrusted{
							SupportedStorageDrivers: []api.DevLXDServerStorageDriverInfo{
								{Name: test.StorageDriver, Remote: test.StorageDriver == "cephfs"},
							},
						},
					}, nil
				},
				CreateVolFunc: func(pool string, volume api.DevLXDStorageVolumesPost) (lxdClient.DevLXDOperation, error) {
					calledCreate = true
					return &devlxd.FakeOperation{}, nil
				},
			}

			controller := NewControllerServer(d)

			req := &csi.CreateVolumeRequest{
				Name: "pvc-8722b28c-a1b2-c3d4-e5f6-a7b8c9d0e1f2",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 67108864, // 64Mi
				},
				VolumeCapabilities: []*csi.VolumeCapability{
					newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, false),
				},
				Parameters: map[string]string{
					ParameterStoragePool: "pool",
				},
			}

			resp, err := controller.CreateVolume(context.Background(), req)
			if test.expectError == "" {
				require.NoError(t, err)
				require.Equal(t, test.StorageDriver, resp.Volume.VolumeContext[ParameterStorageDriver])
				require.True(t, calledCreate, "CreateStoragePoolVolume should have been called")
			} else {
				require.Nil(t, resp)
				require.Equal(t, codes.InvalidArgument, status.Code(err))
				require.ErrorContains(t, err, test.expectError)
				require.False(t, calledCreate, "CreateStoragePoolVolume should not have been called")
			}
		})
	}
}

func TestControllerPublishVolumeRejectsUnsupportedAccessMode(t *testing.T) {
	d := &Driver{
		name:     "lxd.csi.canonical.com",
		version:  "test",
		endpoint: "unix:///csi/csi.sock",
		nodeID:   "test-node",
	}

	d.devLXD = &devlxd.FakeServer{}

	controller := NewControllerServer(d)

	req := &csi.ControllerPublishVolumeRequest{
		VolumeId:         "pool/pvc-volume-name",
		NodeId:           "test-node",
		VolumeCapability: newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, false),
		VolumeContext: map[string]string{
			ParameterStorageDriver: "dir",
		},
	}

	resp, err := controller.ControllerPublishVolume(context.Background(), req)
	require.Nil(t, resp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.ErrorContains(t, err, `Access mode "MULTI_NODE_MULTI_WRITER" is not supported by storage driver "dir"`)
}

func TestControllerPublishVolumeReadonly(t *testing.T) {
	tests := []struct {
		Name           string
		StorageDriver  string
		AccessMode     csi.VolumeCapability_AccessMode_Mode
		Readonly       bool
		expectReadonly string
	}{
		{
			Name:           "Ensure single node writer volume is attached read-write",
			StorageDriver:  "dir",
			AccessMode:     csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			Readonly:       false,
			expectReadonly: "",
		},
		{
			Name:           "Ensure read-only publish request attaches volume read-only",
			StorageDriver:  "dir",
			AccessMode:     csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			Readonly:       true,
			expectReadonly: "true",
		},
		{
			Name:           "Ensure single node reader only volume is attached read-only",
			StorageDriver:  "dir",
			AccessMode:     csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
			Readonly:       false,
			expectReadonly: "true",
		},
		{
			Name:           "Ensure multi node reader only volume is attached read-only on cephfs driver",
			StorageDriver:  "cephfs",
			AccessMode:     csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
			Readonly:       false,
			expectReadonly: "true",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			d := &Driver{
				name:     "lxd.csi.canonical.com",
				version:  "test",
				endpoint: "unix:///csi/csi.sock",
				nodeID:   "test-node",
			}

			var calledUpdate bool
			d.devLXD = &devlxd.FakeServer{
				GetInstFunc: func(name string) (*api.DevLXDInstance, string, error) {
					return &api.DevLXDInstance{Name: name}, "test-etag", nil
				},
				UpdateInstFunc: func(name string, inst api.DevLXDInstancePut, ETag string) error {
					calledUpdate = true
					require.Equal(t, test.expectReadonly, inst.Devices["pvc-volume-name"]["readonly"])
					return nil
				},
			}

			controller := NewControllerServer(d)

			req := &csi.ControllerPublishVolumeRequest{
				VolumeId:         "pool/pvc-volume-name",
				NodeId:           "test-node",
				VolumeCapability: newVolumeCapability(test.AccessMode, false),
				Readonly:         test.Readonly,
				VolumeContext: map[string]string{
					ParameterStorageDriver: test.StorageDriver,
				},
			}

			resp, err := controller.ControllerPublishVolume(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.True(t, calledUpdate, "UpdateInstance should have been called")
		})
	}
}

func TestControllerExpandVolumePreservesConfig(t *testing.T) {
	// Initialize driver and controller server
	d := &Driver{
		name:     "lxd.csi.canonical.com",
		version:  "test",
		endpoint: "unix:///csi/csi.sock",
		nodeID:   "test-node",
	}

	// Create our fake LXD client
	var calledGet, calledUpdate bool
	initialConfig := map[string]string{
		"size":             "21474836480", // 20Gi
		"block.filesystem": "ext4",
		"other.custom.key": "some-value",
	}

	fakeClient := &devlxd.FakeServer{
		GetPoolFunc: func(pool string) (*api.DevLXDStoragePool, string, error) {
			require.Equal(t, "remote", pool)
			return &api.DevLXDStoragePool{
				Name:   pool,
				Driver: "ceph",
			}, "", nil
		},
		GetVolFunc: func(pool string, volType string, name string) (*api.DevLXDStorageVolume, string, error) {
			calledGet = true
			require.Equal(t, "remote", pool)
			require.Equal(t, "custom", volType)
			require.Equal(t, "pvc-volume-name", name)
			return &api.DevLXDStorageVolume{
				Name:        "pvc-volume-name",
				Type:        "custom",
				Description: "Initial description",
				Config:      maps.Clone(initialConfig),
			}, "test-etag", nil
		},
		UpdateVolFunc: func(pool string, volType string, name string, volume api.DevLXDStorageVolumePut, ETag string) (lxdClient.DevLXDOperation, error) {
			calledUpdate = true
			require.Equal(t, "remote", pool)
			require.Equal(t, "custom", volType)
			require.Equal(t, "pvc-volume-name", name)
			require.Equal(t, "test-etag", ETag)
			require.Equal(t, "Initial description", volume.Description)

			// Assert that size is updated and block.filesystem and other keys are preserved
			require.Equal(t, "32212254720", volume.Config["size"]) // 30Gi
			require.Equal(t, "ext4", volume.Config["block.filesystem"])
			require.Equal(t, "some-value", volume.Config["other.custom.key"])
			return &devlxd.FakeOperation{}, nil
		},
	}

	// Inject the fake client directly into the driver
	d.devLXD = fakeClient

	controller := NewControllerServer(d)

	// Invoke ControllerExpandVolume
	req := &csi.ControllerExpandVolumeRequest{
		VolumeId: "remote/pvc-volume-name",
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: 32212254720, // 30Gi
		},
		VolumeCapability: &csi.VolumeCapability{
			AccessMode: &csi.VolumeCapability_AccessMode{
				Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
			},
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType: "ext4",
				},
			},
		},
	}

	resp, err := controller.ControllerExpandVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int64(32212254720), resp.CapacityBytes)

	require.True(t, calledGet, "GetStoragePoolVolume should have been called")
	require.True(t, calledUpdate, "UpdateStoragePoolVolume should have been called")
}
