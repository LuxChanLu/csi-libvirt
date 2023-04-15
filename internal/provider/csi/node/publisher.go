package node

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *Node) NodePublishVolume(ctx context.Context, request *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	n.Logger.Info("publish volume", zap.String("from", request.StagingTargetPath), zap.String("to", request.TargetPath))
	if err := n.Mounter.Mount(request.StagingTargetPath, request.TargetPath, "", []string{"bind"}); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to bind mount: %s", err.Error()))
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

func (n *Node) NodeUnpublishVolume(ctx context.Context, request *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	n.Logger.Info("unpublish volume", zap.String("path", request.TargetPath))
	if err := n.Mounter.Unmount(request.TargetPath); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to bind unmount: %s", err.Error()))
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}
