//go:build integration

package hypervisor_test

import (
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestProvideLibvirt(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ProvideConfig(logger)
	lc := fxtest.NewLifecycle(t)
	libvirt := provider.ProvideLibvirt(lc, logger, cfg)
	lc.RequireStart()
	defer lc.RequireStop()
	assert.NotNil(t, libvirt)
}
