package controller

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
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
	zone := ""
	if request.AccessibilityRequirements.Preferred != nil {
		for _, topology := range request.AccessibilityRequirements.Preferred {
			if zoneSegment, ok := topology.Segments[c.Driver.Name+"/zone"]; ok && zoneSegment != "" {
				zone = zoneSegment
			}
		}
	}
	poolName := request.Parameters["pool"]
	bus := request.Parameters["bus"]
	fstype := request.Parameters["fstype"]
	unlock := c.Driver.DiskLock(poolName, request.Name)
	defer unlock()

	lv, err := c.Hypervisors.Zone(zone)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get libvirt instance: %s, %s", zone, err.Error()))
	}
	pool, err := lv.StoragePoolLookupByName(poolName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool: %s, %s", poolName, err.Error()))
	}
	c.Logger.Info("gonna create a storage volume", zap.String("pool", pool.Name))

	var vol libvirt.StorageVol

	vol, err = lv.StorageVolLookupByName(pool, request.Name)
	if err != nil {
		vol, err = lv.StorageVolCreateXML(pool, c.Driver.Template("volume.xml.tpl", map[string]interface{}{
			"Name":       request.Name,
			"Allocation": request.CapacityRange.LimitBytes,
			"Capacity":   request.CapacityRange.RequiredBytes,
		}), 0)
		if err != nil {
			return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to create volume: %s", err.Error()))
		}
		c.Logger.Info("volume has been created", zap.String("pool", vol.Pool), zap.String("name", vol.Name), zap.String("key", vol.Key))
	} else {
		c.Logger.Info("volume already existing skip creation", zap.String("pool", vol.Pool), zap.String("name", vol.Name), zap.String("key", vol.Key))
	}

	var serial = request.Name
	// TODO: Not the optimal, but some bus shrink the disk serial
	switch bus {
	case "virtio":
		serial = strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(serial))), 16)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      buildVolId(vol.Pool, vol.Name, vol.Key, serial, zone),
			CapacityBytes: request.CapacityRange.RequiredBytes,
			VolumeContext: map[string]string{
				c.Driver.Name + "/pool":   vol.Pool,
				c.Driver.Name + "/bus":    bus,
				c.Driver.Name + "/serial": serial,
				c.Driver.Name + "/fstype": fstype,
				c.Driver.Name + "/zone":   zone,
			},
		},
	}, nil
}

func (c *Controller) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if request.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	poolName, name, key, _, zone := extratVolId(request.VolumeId)
	c.Logger.Info("gonna destroy volume", zap.String("pool", poolName), zap.String("name", name), zap.String("key", key))
	unlock := c.Driver.DiskLock(poolName, name)
	defer unlock()

	lv, err := c.Hypervisors.Zone(zone)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get libvirt instance: %s, %s", zone, err.Error()))
	}
	pool, err := lv.StoragePoolLookupByName(poolName)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unable to get storage pool: %s, %s", poolName, err.Error()))
	}

	vol, err := lv.StorageVolLookupByName(pool, name)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get storage volume: %s, %s", key, err.Error()))
	}

	if err := lv.StorageVolDelete(vol, libvirt.StorageVolDeleteNormal); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to delete storage volume: %s, %s", key, err.Error()))
	}

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

func buildVolId(pool, name, key, serial, zone string) string {
	return strings.Join([]string{pool, name, key, serial}, ":")
}

func extratVolId(volId string) (pool, name, key, serial, zone string) {
	ids := strings.Split(volId, ":")
	if len(ids) < 5 {
		return ids[0], ids[1], ids[2], ids[3], ""
	}
	return ids[0], ids[1], ids[2], ids[3], ids[4]
}
