///go:build integration

package controller_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type testController struct {
	csi.ControllerServer
	t       *testing.T
	libvirt *libvirt.Libvirt
	driver  *driver.Driver
}

func testContoller(t *testing.T) *testController {
	logger := zap.NewNop()
	config := config.ProvideConfig(logger)
	lc := fxtest.NewLifecycle(t)
	libvirt := provider.ProvideLibvirt(lc, logger, config)
	lc.RequireStart()
	t.Cleanup(lc.RequireStop)
	driver := driver.ProvideDriver(config, logger)
	return &testController{libvirt: libvirt, t: t, driver: driver, ControllerServer: provider.ProvideCSIController(driver, logger, libvirt)}
}

func (tc *testController) createTestVolume(name string) (*csi.CapacityRange, *csi.CreateVolumeResponse) {
	capacity := &csi.CapacityRange{RequiredBytes: 1024 * 1024 * 50, LimitBytes: 1024 * 1024 * 100}
	response, err := tc.CreateVolume(context.Background(), &csi.CreateVolumeRequest{
		Name:          name,
		CapacityRange: capacity,
		VolumeCapabilities: []*csi.VolumeCapability{
			{AccessType: &csi.VolumeCapability_Block{Block: &csi.VolumeCapability_BlockVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}},
			{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{FsType: "ext4"}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}},
		},
		Parameters: map[string]string{
			"pool":   "default",
			"bus":    "scsi",
			"fstype": "ext4",
		},
	})
	assert.NoError(tc.t, err)
	assert.NotEmpty(tc.t, response.Volume.VolumeId)
	assert.Equal(tc.t, response.Volume.VolumeId, fmt.Sprintf("default:%s:/var/lib/libvirt/images/%s:%s", name, name, name))
	return capacity, response
}

func TestControllerGetCapabilities(t *testing.T) {
	controller := provider.ProvideCSIController(&driver.Driver{}, zap.NewNop(), nil)
	response, err := controller.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Capabilities)
}

func TestControllerExpandVolume(t *testing.T) {
	controller := testContoller(t)
	name := "pvc-to-expand"
	capacity, createResponse := controller.createTestVolume(name)
	capacity.LimitBytes *= 2
	capacity.RequiredBytes *= 2
	response, err := controller.ControllerExpandVolume(context.Background(), &csi.ControllerExpandVolumeRequest{
		VolumeId: createResponse.Volume.VolumeId, CapacityRange: capacity,
		VolumeCapability: &csi.VolumeCapability{AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}, AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{FsType: "ext4"}}},
	})
	assert.NoError(t, err)
	assert.Equal(t, capacity.LimitBytes, response.CapacityBytes)
	assert.True(t, response.NodeExpansionRequired)

	// Assert actually resized in libvirt
	pool, err := controller.libvirt.StoragePoolLookupByName("default")
	assert.NoError(t, err)
	vol, err := controller.libvirt.StorageVolLookupByName(pool, name)
	assert.NoError(t, err)
	volType, volCapacity, volAllocation, err := controller.libvirt.StorageVolGetInfo(vol)
	assert.NoError(t, err)
	assert.Equal(t, int64(209719296), int64(volAllocation))
	assert.Equal(t, response.CapacityBytes, int64(volCapacity))
	assert.Equal(t, volType, int8(0))
	assert.NoError(t, controller.libvirt.StorageVolDelete(vol, 0))
}

func TestControllerGetVolume(t *testing.T) {
	controller := testContoller(t)
	name := "pvc-to-get"
	capacity, createResponse := controller.createTestVolume(name)
	response, err := controller.ControllerGetVolume(context.Background(), &csi.ControllerGetVolumeRequest{VolumeId: createResponse.Volume.VolumeId})
	assert.NoError(t, err)
	assert.Equal(t, capacity.LimitBytes, response.Volume.CapacityBytes)
}
