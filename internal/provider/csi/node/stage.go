package node

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

func (n *Node) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	serial := request.PublishContext[n.Driver.Name+"/serial"]
	dev := request.PublishContext[n.Driver.Name+"/dev"]
	fstype := request.VolumeContext[n.Driver.Name+"/fstype"]
	n.Logger.Info("gonna format/mount (if necessary)", zap.String("serial", serial), zap.String("dev", dev), zap.String("fstype", fstype))
	disks, err := filepath.Glob("/dev/disk/by-id/*" + serial)
	if err != nil || len(disks) != 1 {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to find attached disk: %s", err.Error()))
	}
	n.Logger.Info("source disk is", zap.String("disk", disks[0]))
	if err := n.Formatter.FormatAndMount(disks[0], request.StagingTargetPath, fstype, []string{"rw"}); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to format/mount attached disk: %s", err.Error()))
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

func (n *Node) NodeUnstageVolume(ctx context.Context, request *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	n.Logger.Info("gonna umount", zap.String("path", request.StagingTargetPath), zap.String("volId", request.VolumeId))
	if exist, err := mount.PathExists(request.StagingTargetPath); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to check if path exist: %s", err.Error()))
	} else if !exist {
		n.Logger.Info("skip umount folder not existing", zap.String("path", request.StagingTargetPath), zap.String("volId", request.VolumeId))
		return &csi.NodeUnstageVolumeResponse{}, nil
	}
	if err := n.Formatter.Unmount(request.StagingTargetPath); err != nil {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("unable to unmount attached disk: %s", err.Error()))
	}
	return &csi.NodeUnstageVolumeResponse{}, nil
}
