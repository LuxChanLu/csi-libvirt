package controller

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var devPrefixes = map[string]string{
	"virtio": "vd",
	"usb":    "sd", "scsi": "sd", "sata": "sd",
	"ide": "hd",
}

func (c *Controller) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	var nodeUuid = libvirt.UUID{}
	nodeUuidRaw, err := hex.DecodeString(request.NodeId)
	if err != nil || copy(nodeUuid[:], nodeUuidRaw) != libvirt.UUIDBuflen {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to parse uuid node id: %s", err.Error()))
	}
	poolName, name, key, _ := extratVolId(request.VolumeId)
	unlock := c.Driver.DiskLock(poolName, name)
	defer unlock()

	c.Logger.Info("volume gonna be published", zap.String("nodeId", request.NodeId), zap.String("pool", poolName), zap.String("name", name), zap.String("key", key))
	domain, err := c.Libvirt.DomainLookupByUUID(nodeUuid)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	bus := request.VolumeContext[c.Driver.Name+"/bus"]
	serial := request.VolumeContext[c.Driver.Name+"/serial"]
	devPrefix := devPrefixes[bus]
	dev, alreadyMounted, err := c.genDiskTargetSuffix(domain, devPrefix, key)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to generate disk target: %s", err.Error()))
	}
	if alreadyMounted {
		c.Logger.Info("volume already attached to domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name), zap.String("serial", serial), zap.String("bus", bus), zap.String("dev", dev))
	} else if err := c.Libvirt.DomainAttachDevice(domain, c.Driver.Template("disk.xml.tpl", map[string]interface{}{"Source": key, "Bus": bus, "Dev": dev, "Serial": serial})); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to attach disk to domain: %s", err.Error()))
	} else if err == nil {
		c.Logger.Info("volume attached to domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name), zap.String("serial", serial), zap.String("bus", bus), zap.String("dev", dev))
	}
	alias, err := c.getDiskAlias(domain, serial)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain disk alias: %s", err.Error()))
	}
	return &csi.ControllerPublishVolumeResponse{PublishContext: map[string]string{c.Driver.Name + "/serial": serial, c.Driver.Name + "/dev": dev, c.Driver.Name + "/alias": alias}}, nil
}

func (c *Controller) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	var nodeUuid = libvirt.UUID{}
	nodeUuidRaw, err := hex.DecodeString(request.NodeId)
	if err != nil || copy(nodeUuid[:], nodeUuidRaw) != libvirt.UUIDBuflen {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to parse uuid node id: %s", err.Error()))
	}
	poolName, name, key, serial := extratVolId(request.VolumeId)
	unlock := c.Driver.DiskLock(poolName, name)
	defer unlock()

	c.Logger.Info("volume gonna be unpublished", zap.String("nodeId", request.NodeId), zap.String("pool", poolName), zap.String("name", name), zap.String("key", key))
	domain, err := c.Libvirt.DomainLookupByUUID(nodeUuid)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	alias, err := c.getDiskAlias(domain, serial)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get disk alias: %s", err.Error()))
	}
	if err := c.Libvirt.DomainDetachDeviceAlias(domain, alias, 0); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to dettached disk to domain: %s", err.Error()))
	}
	c.Logger.Info("volume dettached from domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name))
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (c *Controller) genDiskTargetSuffix(domain libvirt.Domain, prefix, source string) (string, bool, error) {
	xml, err := c.Libvirt.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return "", false, err
	}
	disks, err := c.Driver.LookupDomainDisks(xml)
	if err != nil {
		return "", false, err
	}
	for _, disk := range disks {
		if strings.EqualFold(disk.Source.File, source) {
			return disk.Target.Dev, true, nil
		}
	}
	idx := 1
	dev := ""
	for {
		dev = fmt.Sprintf("%s%s", prefix, c.Driver.EncodeNumberToAlphabet(idx))
		for _, disk := range disks {
			if strings.EqualFold(disk.Target.Dev, dev) {
				dev = ""
			}
		}
		if dev != "" {
			return dev, false, nil
		}
		idx++
	}
}

func (c *Controller) getDiskAlias(domain libvirt.Domain, serial string) (string, error) {
	xml, err := c.Libvirt.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return "", err
	}
	disks, err := c.Driver.LookupDomainDisks(xml)
	if err != nil {
		return "", err
	}
	for _, disk := range disks {
		if strings.EqualFold(disk.Serial, serial) {
			return disk.Alias.Name, nil
		}
	}
	return "", fmt.Errorf("unable to find disk %s", serial)
}
