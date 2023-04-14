package node

import (
	"context"

	"github.com/LuxChanLu/libvirt-csi/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
)

type Node struct {
	Driver *driver.Driver
}

func (n *Node) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}

func (n *Node) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, nil
}

func (n *Node) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, nil
}
