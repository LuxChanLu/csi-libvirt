package controller_test

import (
	"context"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/controller"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
)

// func TestControllerPublishVolume(t *testing.T) {
// 	controller := testContoller(t)
// 	_, createResponse := controller.createTestVolume("pvc-to-publish")
// 	domain := test.CreateTestDomain(t, controller.libvirt)
// 	response, err := controller.ControllerPublishVolume(context.Background(), &csi.ControllerPublishVolumeRequest{
// 		VolumeId: createResponse.Volume.VolumeId,
// 		NodeId:   hex.EncodeToString(domain.UUID[:]),
// 		VolumeContext: map[string]string{
// 			controller.driver.Name + "/bus":    "scsi",
// 			controller.driver.Name + "/serial": "pvc-to-publish",
// 		},
// 	})
// 	assert.Error(t, err)
// 	assert.Equal(t, "", response.PublishContext[controller.driver.Name+"/dev"])
// 	assert.Equal(t, "", response.PublishContext[controller.driver.Name+"/alias"])
// }

func TestControllerUnpublishVolume(t *testing.T) {
	controller := &controller.Controller{}
	_, err := controller.DeleteSnapshot(context.Background(), &csi.DeleteSnapshotRequest{})
	assert.Error(t, err)
}
