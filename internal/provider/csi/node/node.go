package node

import (
	"context"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"k8s.io/utils/mount"
)

type Node struct {
	Driver    *driver.Driver
	Logger    *zap.Logger
	MachineID string
	Formatter *mount.SafeFormatAndMount
	Mounter   mount.Interface
	Zone      string
}

func (n *Node) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}

func (n *Node) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
	}, nil
}

func (n *Node) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	n.Logger.Info("NodeGetInfo called")
	response := &csi.NodeGetInfoResponse{NodeId: n.MachineID}
	if n.Zone != "" {
		response.AccessibleTopology = &csi.Topology{Segments: map[string]string{n.Driver.Name + "/zone": n.Zone}}
	}
	return response, nil
}
