package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (c *Controller) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	if request.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "ControllerExpandVolume Name must be provided")
	}

	if request.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "ControllerExpandVolume Volume VolumeCapability must be provided")
	} else if violations := validateCapabilities([]*csi.VolumeCapability{request.VolumeCapability}); len(violations) > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume capabilities cannot be satisified: %s", strings.Join(violations, "; ")))
	}
	poolName, name, _, _ := extratVolId(request.VolumeId)

	pool, err := c.Libvirt.StoragePoolLookupByName(poolName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool: %s, %s", poolName, err.Error()))
	}
	vol, err := c.Libvirt.StorageVolLookupByName(pool, name)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool volume: %s, %s", name, err.Error()))
	}
	err = c.Libvirt.StorageVolResize(vol, uint64(request.CapacityRange.RequiredBytes), 0)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to resize volume: %s, %s", name, err.Error()))
	}
	err = c.Libvirt.StorageVolResize(vol, uint64(request.CapacityRange.LimitBytes), libvirt.StorageVolResizeAllocate)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to resize allocated volume: %s, %s", name, err.Error()))
	}
	_, volCapacity, _, err := c.Libvirt.StorageVolGetInfo(vol)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get volume information: %s", err.Error()))
	}

	return &csi.ControllerExpandVolumeResponse{NodeExpansionRequired: true, CapacityBytes: int64(volCapacity)}, nil
}

func (c *Controller) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	poolName, name, key, _ := extratVolId(request.VolumeId)
	response := &csi.ControllerGetVolumeResponse{Volume: &csi.Volume{
		VolumeId:      request.VolumeId,
		VolumeContext: map[string]string{},
	}, Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
		PublishedNodeIds: []string{},
		VolumeCondition:  &csi.VolumeCondition{Abnormal: false, Message: "Disk OK"},
	}}
	pool, err := c.Libvirt.StoragePoolLookupByName(poolName)
	if err != nil {
		response.Status.VolumeCondition = &csi.VolumeCondition{Abnormal: true, Message: fmt.Sprintf("Pool %s not found or with error (%s)", poolName, err.Error())}
		return response, nil
	}
	vol, err := c.Libvirt.StorageVolLookupByName(pool, name)
	if err != nil {
		response.Status.VolumeCondition = &csi.VolumeCondition{Abnormal: true, Message: fmt.Sprintf("Disk %s in pool %s not found or with error (%s)", name, poolName, err.Error())}
		return response, nil
	}
	_, volCapacity, _, err := c.Libvirt.StorageVolGetInfo(vol)
	if err != nil {
		response.Status.VolumeCondition = &csi.VolumeCondition{Abnormal: true, Message: fmt.Sprintf("Unable to get disk %s in pool %s informations (%s)", name, poolName, err.Error())}
		return response, nil
	}
	response.Volume.CapacityBytes = int64(volCapacity)
	nodeIds, err := c.Driver.DiskAttachedToNodes(ctx, key)
	if err != nil {
		return response, nil
	}
	response.Status.PublishedNodeIds = nodeIds
	return response, nil
}
