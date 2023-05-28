package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
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
	totalDiskCount := 0
	allHypervisorsDisks := make([][][]*csi.ListVolumesResponse_Entry, len(c.Hypervisors.Libvirts))
	for hyperisorIdx, lv := range c.Hypervisors.Libvirts {
		pools, _, err := lv.ConnectListAllStoragePools(1, 0)
		if err != nil {
			return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to dettached disk to domain: %s", err.Error()))
		}
		allHypervisorsDisks[hyperisorIdx] = make([][]*csi.ListVolumesResponse_Entry, len(pools))
		for idx, pool := range pools {
			idx := idx
			pool := pool
			vols, _, err := lv.StoragePoolListAllVolumes(pool, 1, 0)
			if err != nil {
				return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to list all volumes: %s", err.Error()))
			}
			allHypervisorsDisks[hyperisorIdx][idx] = make([]*csi.ListVolumesResponse_Entry, len(vols))
			for idxVol, vol := range vols {
				volume, err := c.ControllerGetVolume(context.Background(), &csi.ControllerGetVolumeRequest{VolumeId: buildVolId(vol.Pool, vol.Name, vol.Key, "", lv.Zone)})
				if err != nil {
					return nil, err
				}
				allHypervisorsDisks[hyperisorIdx][idx][idxVol] = &csi.ListVolumesResponse_Entry{Volume: volume.Volume, Status: &csi.ListVolumesResponse_VolumeStatus{PublishedNodeIds: volume.Status.PublishedNodeIds, VolumeCondition: volume.Status.VolumeCondition}}
			}
			totalDiskCount += len(vols)
		}
	}
	disks := make([]*csi.ListVolumesResponse_Entry, totalDiskCount)
	idx := 0
	for _, hypervisorsDisks := range allHypervisorsDisks {
		for _, allDisk := range hypervisorsDisks {
			for _, disk := range allDisk {
				disks[idx] = disk
			}
		}
	}
	startingToken, err := strconv.ParseInt(request.StartingToken, 10, 64)
	if err != nil {
		startingToken = 0
	}
	paginatedDisk := disks[startingToken:request.MaxEntries]
	return &csi.ListVolumesResponse{Entries: paginatedDisk, NextToken: fmt.Sprintf("%d", len(paginatedDisk))}, nil
}

func (c *Controller) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	poolName := request.Parameters["pool"]
	totalRAvailable := uint64(0)
	for _, lv := range c.Hypervisors.Libvirts {
		pool, err := lv.StoragePoolLookupByName(poolName)
		if err != nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool: %s, %s", poolName, err.Error()))
		}
		_, _, _, rAvailable, err := lv.StoragePoolGetInfo(pool)
		if err != nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool information: %s, %s", poolName, err.Error()))
		}
		totalRAvailable += rAvailable
	}
	return &csi.GetCapacityResponse{AvailableCapacity: int64(totalRAvailable)}, nil
}
