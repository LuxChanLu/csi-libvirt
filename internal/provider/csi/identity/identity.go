package identity

import (
	"context"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Identity struct {
	Driver *driver.Driver
}

func (i *Identity) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{Name: i.Driver.Name, VendorVersion: i.Driver.Version}, nil
}

func (i *Identity) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_OFFLINE,
					},
				},
			},
		},
	}, nil
}

func (i *Identity) Probe(ctx context.Context, request *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	if i.Driver.Hypervisors != nil {
		for _, lv := range i.Driver.Hypervisors.Libvirts {
			_, err := lv.ConnectGetLibVersion()
			if err != nil {
				return &csi.ProbeResponse{Ready: wrapperspb.Bool(false)}, nil
			}
		}
	}
	return &csi.ProbeResponse{Ready: wrapperspb.Bool(true)}, nil
}
