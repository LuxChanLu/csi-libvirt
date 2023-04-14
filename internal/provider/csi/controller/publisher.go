package controller

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (c *Controller) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, nil
}

func (c *Controller) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}
