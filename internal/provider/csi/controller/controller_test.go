package controller_test

import (
	"context"
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
}

func testContoller(t *testing.T) *testController {
	logger := zap.NewNop()
	config := config.ProvideConfig(logger)
	lc := fxtest.NewLifecycle(t)
	libvirt := provider.ProvideLibvirt(lc, logger, config)
	lc.RequireStart()
	t.Cleanup(lc.RequireStop)
	return &testController{libvirt: libvirt, t: t, ControllerServer: provider.ProvideCSIController(driver.ProvideDriver(config, logger), logger, libvirt)}
}

func TestControllerGetCapabilities(t *testing.T) {
	controller := provider.ProvideCSIController(&driver.Driver{}, zap.NewNop(), nil)
	response, err := controller.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Capabilities)
}

// func TestControllerGetVolume(t *testing.T) {
// 	logger := zap.NewNop()
// 	lc := fxtest.NewLifecycle(t)
// 	libvirt := provider.ProvideLibvirt(lc, logger, config.ProvideConfig(logger))
// 	pool, err := libvirt.StoragePool("default")
// 	assert.NoError(t, err)
// 	vol, err := libvirt.StorageVolCreateXML(pool, ``, 0)
// 	controller := provider.ProvideCSIController(&driver.Driver{}, logger, libvirt)
// 	response, err := controller.ControllerGetVolume(context.Background(), &csi.ControllerGetVolumeRequest{VolumeId: ""})
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, response)
// }
