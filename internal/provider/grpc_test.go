package provider_test

import (
	"os"
	"syscall"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestProvideCSI(t *testing.T) {
	logger := zap.NewNop()
	lc := fxtest.NewLifecycle(t)
	driver := &driver.Driver{Endpoint: "/tmp/csi-libvirt-test.sock"}
	grpcServer := provider.ProvideGRPCServer(lc, driver, logger)
	lc.RequireStart()
	defer lc.RequireStop()
	assert.NotNil(t, grpcServer)
	stat, err := os.Stat(driver.Endpoint)
	assert.NoError(t, err)
	assert.NotNil(t, stat)
	assert.Equal(t, uint32(syscall.S_IFSOCK), stat.Sys().(*syscall.Stat_t).Mode&syscall.S_IFMT)
}
