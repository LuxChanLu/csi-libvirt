package controller

import (
	"context"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
)

type Controller struct {
	Driver  *driver.Driver
	Libvirt *libvirt.Libvirt
	Logger  *zap.Logger
}

func (c *Controller) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	cap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			cap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME),
			cap(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME),
			cap(csi.ControllerServiceCapability_RPC_LIST_VOLUMES),
			cap(csi.ControllerServiceCapability_RPC_GET_CAPACITY),
			cap(csi.ControllerServiceCapability_RPC_EXPAND_VOLUME),
			cap(csi.ControllerServiceCapability_RPC_GET_VOLUME),
			cap(csi.ControllerServiceCapability_RPC_LIST_VOLUMES_PUBLISHED_NODES),
		},
	}, nil
}

func (c *Controller) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

func (c *Controller) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, nil
}
