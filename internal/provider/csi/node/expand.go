package node

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (n *Node) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}
