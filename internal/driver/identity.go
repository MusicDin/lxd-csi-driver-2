package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type identityServer struct {
	driver *Driver

	// Must be embedded for forward compatibility.
	csi.UnimplementedIdentityServer
}

// NewIdentityServer returns a new instance of the CSI identity server.
func NewIdentityServer(driver *Driver) *identityServer {
	return &identityServer{
		driver: driver,
	}
}

// GetPluginInfo retrieves the plugin information.
func (i *identityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	if i.driver.name == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing name")
	}

	if i.driver.version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          i.driver.name,
		VendorVersion: i.driver.version,
	}, nil
}

// GetPluginCapabilities retrieves the plugin capabilities.
func (i *identityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}
