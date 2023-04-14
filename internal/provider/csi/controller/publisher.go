package controller

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Controller) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	var nodeUuid = libvirt.UUID{}
	nodeUuidRaw, err := hex.DecodeString(request.NodeId)
	if err != nil || copy(nodeUuid[:], nodeUuidRaw) != libvirt.UUIDBuflen {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to parse uuid node id: %s", err.Error()))
	}
	poolName, name, key := extratVolId(request.VolumeId)
	c.Logger.Info("volume gonna be published", zap.String("nodeId", request.NodeId), zap.String("pool", poolName), zap.String("name", name), zap.String("key", key))
	domain, err := c.Libvirt.DomainLookupByUUID(nodeUuid)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	if err := c.Libvirt.DomainAttachDevice(domain, c.Driver.Template("disk.xml.tpl", map[string]interface{}{
		"Alias":  name,
		"Source": key,
		"Bus":    request.VolumeContext[c.Driver.Name+"/bus"],
	})); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to attach disk to domain: %s", err.Error()))
	}
	c.Logger.Info("volume attached to domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name))
	return nil, nil
}

func (c *Controller) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	var nodeUuid = libvirt.UUID{}
	nodeUuidRaw, err := hex.DecodeString(request.NodeId)
	if err != nil || copy(nodeUuid[:], nodeUuidRaw) != libvirt.UUIDBuflen {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to parse uuid node id: %s", err.Error()))
	}
	poolName, name, key := extratVolId(request.VolumeId)
	c.Logger.Info("volume gonna be unpublished", zap.String("nodeId", request.NodeId), zap.String("pool", poolName), zap.String("name", name), zap.String("key", key))
	domain, err := c.Libvirt.DomainLookupByUUID(nodeUuid)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to get domain: %s", err.Error()))
	}
	if err := c.Libvirt.DomainDetachDeviceAlias(domain, name, 0); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to attach disk to domain: %s", err.Error()))
	}
	c.Logger.Info("volume dettached from domain", zap.String("nodeId", request.NodeId), zap.String("volId", request.VolumeId), zap.String("domain", domain.Name))
	return nil, nil
}
