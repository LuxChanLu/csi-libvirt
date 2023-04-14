package controller

import (
	"context"

	"github.com/LuxChanLu/libvirt-csi/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
)

type Controller struct {
	Driver  *driver.Driver
	Libvirt *libvirt.Libvirt
}

func (c *Controller) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return nil, nil
}

func (c *Controller) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

func (c *Controller) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, nil
}
