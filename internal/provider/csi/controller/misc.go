package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Controller) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	if violations := validateCapabilities(request.VolumeCapabilities); len(violations) > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume capabilities cannot be satisified: %s", strings.Join(violations, "; ")))
	}
	return &csi.ValidateVolumeCapabilitiesResponse{Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
		VolumeContext:      request.VolumeContext,
		VolumeCapabilities: request.VolumeCapabilities,
		Parameters:         request.Parameters,
	}, Message: "Disk OK"}, nil
}

func (c *Controller) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	pools, _, err := c.Libvirt.ConnectListAllStoragePools(1, 0)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to dettached disk to domain: %s", err.Error()))
	}
	wg, gCtx := errgroup.WithContext(ctx)
	allDisks := make([][]*csi.ListVolumesResponse_Entry, len(pools))
	totalDiskCount := 0
	for idx, pool := range pools {
		idx := idx
		pool := pool
		wg.Go(func() error {
			vols, _, err := c.Libvirt.StoragePoolListAllVolumes(pool, 1, 0)
			if err != nil {
				return err
			}
			allDisks[idx] = make([]*csi.ListVolumesResponse_Entry, len(vols))
			for idxVol, vol := range vols {
				volume, err := c.ControllerGetVolume(gCtx, &csi.ControllerGetVolumeRequest{VolumeId: buildVolId(vol.Pool, vol.Name, vol.Key, "")})
				if err != nil {
					return err
				}
				allDisks[idx][idxVol] = &csi.ListVolumesResponse_Entry{Volume: volume.Volume, Status: &csi.ListVolumesResponse_VolumeStatus{PublishedNodeIds: volume.Status.PublishedNodeIds, VolumeCondition: volume.Status.VolumeCondition}}
			}
			totalDiskCount += len(vols)
			return nil
		})
	}
	disks := make([]*csi.ListVolumesResponse_Entry, totalDiskCount)
	idx := 0
	for _, allDisk := range allDisks {
		for _, disk := range allDisk {
			disks[idx] = disk
		}
	}
	startingToken, err := strconv.ParseInt(request.StartingToken, 10, 64)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to parse starting token: %s", err.Error()))
	}
	paginatedDisk := disks[startingToken:request.MaxEntries]
	return &csi.ListVolumesResponse{Entries: paginatedDisk, NextToken: fmt.Sprintf("%d", len(paginatedDisk))}, nil
}

func (c *Controller) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	poolName := request.Parameters["pool"]
	pool, err := c.Libvirt.StoragePoolLookupByName(poolName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool: %s, %s", poolName, err.Error()))
	}
	_, _, _, rAvailable, err := c.Libvirt.StoragePoolGetInfo(pool)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool information: %s, %s", poolName, err.Error()))
	}
	return &csi.GetCapacityResponse{AvailableCapacity: int64(rAvailable)}, nil
}
