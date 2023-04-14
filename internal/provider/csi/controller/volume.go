package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	supportedAccessMode = &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
)

func (c *Controller) CreateVolume(ctx context.Context, request *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if request.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	if request.VolumeCapabilities == nil || len(request.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Volume capabilities must be provided")
	}

	if violations := validateCapabilities(request.VolumeCapabilities); len(violations) > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("volume capabilities cannot be satisified: %s", strings.Join(violations, "; ")))
	}

	pool, err := c.Libvirt.StoragePool(request.Parameters["pool"])
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("storage pool not found: %s, %s", request.Parameters["pool"], err.Error()))
	}

	var vol libvirt.StorageVol

	vol, err = c.Libvirt.StorageVolLookupByName(pool, request.Name)
	if err != nil {
		vol, err = c.Libvirt.StorageVolCreateXML(pool, c.Driver.Template("volume.xml.tpl", map[string]interface{}{
			"Name":       request.Name,
			"Allocation": request.CapacityRange.LimitBytes,
			"Capacity":   request.CapacityRange.RequiredBytes,
		}), libvirt.StorageVolCreatePreallocMetadata)
		if err != nil {
			return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to create volume: %s", err.Error()))
		}
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      vol.Key,
			CapacityBytes: request.CapacityRange.RequiredBytes,
			VolumeContext: map[string]string{"pool": vol.Pool},
		},
	}, nil
}

func (c *Controller) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if request.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	// pool, err := c.Libvirt.StoragePool(request.Parameters["pool"])
	// if err != nil {
	// 	return nil, status.Error(codes.NotFound, fmt.Sprintf("storage pool not found: %s, %s", request.Parameters["pool"], err.Error()))
	// }

	// vol, err = c.Libvirt.StorageVolLookupByKey(pool, request.VolumeId)
	// c.Libvirt.StorageVolDelete()
	return &csi.DeleteVolumeResponse{}, nil
}

func validateCapabilities(caps []*csi.VolumeCapability) []string {
	violations := sets.NewString()
	for _, cap := range caps {
		if cap.GetAccessMode().GetMode() != supportedAccessMode.GetMode() {
			violations.Insert(fmt.Sprintf("unsupported access mode %s", cap.GetAccessMode().GetMode().String()))
		}

		accessType := cap.GetAccessType()
		switch accessType.(type) {
		case *csi.VolumeCapability_Block:
		case *csi.VolumeCapability_Mount:
		default:
			violations.Insert("unsupported access type")
		}
	}

	return violations.List()
}
