package node

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *Node) NodePublishVolume(ctx context.Context, request *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	n.Logger.Info("publish volume", zap.String("from", request.StagingTargetPath), zap.String("to", request.TargetPath))
	err := os.MkdirAll(filepath.Dir(request.TargetPath), 0750)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("failed to create target directory for raw block bind mount: %v", err))
	}

	file, err := os.OpenFile(request.TargetPath, os.O_CREATE, 0660)
	if err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("failed to create target file for raw block bind mount: %v", err))
	}
	file.Close()
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
	if err := os.Remove(request.TargetPath); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to cleanup mount path: %s", err.Error()))
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}
