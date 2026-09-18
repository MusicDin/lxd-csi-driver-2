package driver

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/require"
)

// newVolumeCapability returns a VolumeCapability with the given access mode and
// either block or mount access type.
func newVolumeCapability(mode csi.VolumeCapability_AccessMode_Mode, block bool) *csi.VolumeCapability {
	volCap := &csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: mode,
		},
	}

	if block {
		volCap.AccessType = &csi.VolumeCapability_Block{
			Block: &csi.VolumeCapability_BlockVolume{},
		}
	} else {
		volCap.AccessType = &csi.VolumeCapability_Mount{
			Mount: &csi.VolumeCapability_MountVolume{},
		}
	}

	return volCap
}

func TestValidateVolumeCapabilities(t *testing.T) {
	tests := []struct {
		Name               string
		StorageDriver      string
		VolumeCapabilities []*csi.VolumeCapability
		expectError        string
	}{
		{
			Name:          "Ensure single node writer is accepted on local driver",
			StorageDriver: "dir",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER, false),
			},
			expectError: "",
		},
		{
			Name:          "Ensure single node single writer is accepted on remote driver",
			StorageDriver: "ceph",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER, false),
			},
			expectError: "",
		},
		{
			Name:          "Ensure single node writer is accepted for block volume",
			StorageDriver: "zfs",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER, true),
			},
			expectError: "",
		},
		{
			Name:          "Ensure unset access mode is accepted",
			StorageDriver: "dir",
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
			},
			expectError: "",
		},
		{
			Name:          "Ensure multi node multi writer is accepted on cephfs driver",
			StorageDriver: "cephfs",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, false),
			},
			expectError: "",
		},
		{
			Name:          "Ensure multi node single writer is accepted on cephfs driver",
			StorageDriver: "cephfs",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER, false),
			},
			expectError: "",
		},
		{
			Name:          "Ensure multi node reader only is accepted on cephfs driver",
			StorageDriver: "cephfs",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY, false),
			},
			expectError: "",
		},
		{
			Name:          "Ensure multi node multi writer is rejected on local driver",
			StorageDriver: "dir",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, false),
			},
			expectError: `Access mode "MULTI_NODE_MULTI_WRITER" is not supported by storage driver "dir"`,
		},
		{
			Name:          "Ensure multi node reader only is rejected on remote single node driver",
			StorageDriver: "ceph",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY, false),
			},
			expectError: `Access mode "MULTI_NODE_READER_ONLY" is not supported by storage driver "ceph"`,
		},
		{
			Name:          "Ensure multi node single writer is rejected on unknown driver",
			StorageDriver: "",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER, false),
			},
			expectError: `Access mode "MULTI_NODE_SINGLE_WRITER" is not supported by storage driver ""`,
		},
		{
			Name:          "Ensure multi node multi writer is rejected for block volume on cephfs driver",
			StorageDriver: "cephfs",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, true),
			},
			expectError: `Access mode "MULTI_NODE_MULTI_WRITER" is not supported for block volumes`,
		},
		{
			Name:          "Ensure any multi node capability is rejected on local driver",
			StorageDriver: "lvm",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER, false),
				newVolumeCapability(csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER, false),
			},
			expectError: `Access mode "MULTI_NODE_MULTI_WRITER" is not supported by storage driver "lvm"`,
		},
		{
			Name:          "Ensure nil capability is rejected",
			StorageDriver: "dir",
			VolumeCapabilities: []*csi.VolumeCapability{
				newVolumeCapability(csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER, false),
				nil,
			},
			expectError: "VolumeCapability cannot be nil",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := ValidateVolumeCapabilities(test.StorageDriver, test.VolumeCapabilities...)
			if test.expectError == "" {
				require.NoError(t, err, "Expected no error, got %v", err)
			} else {
				if err == nil {
					require.FailNowf(t, "Expected error %q, got none", test.expectError)
				}

				require.ErrorContains(t, err, test.expectError)
			}
		})
	}
}
