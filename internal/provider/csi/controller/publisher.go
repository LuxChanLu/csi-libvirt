package controller

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/beevik/etree"
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

	domain, err := c.Libvirt.DomainLookupByUUID(nodeUuid)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	domainXml, err := c.Libvirt.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	bus := request.VolumeContext[c.Driver.Name+"/bus"]
	serial := request.VolumeContext[c.Driver.Name+"/serial"]
	devPrefix := devPrefixes[bus]
	dev, err := c.genDiskTargetSuffix(domainXml, devPrefix)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to generate disk target: %s", err.Error()))
	}
	c.Logger.Info("volume gonna be published", zap.String("nodeId", request.NodeId), zap.String("pool", poolName), zap.String("name", name), zap.String("key", key), zap.String("dev", dev))
	if err := c.Driver.AttachDisk(domainXml, c.Driver.Template("disk.xml.tpl", map[string]interface{}{"Source": key, "Bus": bus, "Dev": dev, "Serial": serial}), serial); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to attach disk to domain: %s", err.Error()))
	} else {
		c.Logger.Info("volume attached to domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name), zap.String("serial", serial), zap.String("bus", bus), zap.String("dev", dev))
	}
	return &csi.ControllerPublishVolumeResponse{PublishContext: map[string]string{c.Driver.Name + "/serial": serial, c.Driver.Name + "/dev": dev}}, nil
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
	domainXml, err := c.Libvirt.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	if err := c.Driver.DettachDisk(domainXml, serial); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to dettached disk to domain: %s", err.Error()))
	}
	c.Logger.Info("volume dettached from domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name))
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (c *Controller) genDiskTargetSuffix(domainXml, prefix string) (string, error) {
	domainDoc := etree.NewDocument()
	if err := domainDoc.ReadFromString(domainXml); err != nil {
		return "", err
	}
	targetDevs := []string{}
	actualDisks := domainDoc.FindElements("//domain/devices/disk")
	for _, actualDisk := range actualDisks {
		target := actualDisk.FindElement("//target")
		if target != nil {
			dev := target.SelectAttr("dev")
			if dev != nil {
				targetDevs = append(targetDevs, dev.Value)
			}
		}
	}
	dev := ""
	for i := 1; dev == ""; i++ {
		dev = fmt.Sprintf("%s%s", prefix, c.Driver.EncodeNumberToAlphabet(i))
		for _, existingDev := range targetDevs {
			if existingDev == dev {
				dev = ""
				break
			}
		}
	}
	return dev, nil
}
