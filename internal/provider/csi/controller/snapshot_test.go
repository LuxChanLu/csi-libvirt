package controller_test

import (
	"context"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/controller"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshot(t *testing.T) {
	controller := &controller.Controller{}
	_, err := controller.CreateSnapshot(context.Background(), &csi.CreateSnapshotRequest{})
	assert.Error(t, err)
}

func TestDeleteSnapshot(t *testing.T) {
	controller := &controller.Controller{}
	_, err := controller.DeleteSnapshot(context.Background(), &csi.DeleteSnapshotRequest{})
	assert.Error(t, err)
}

func TestListSnapshots(t *testing.T) {
	controller := &controller.Controller{}
	_, err := controller.ListSnapshots(context.Background(), &csi.ListSnapshotsRequest{})
	assert.Error(t, err)
}
