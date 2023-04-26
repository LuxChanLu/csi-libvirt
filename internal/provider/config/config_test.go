//go:build !integration

package config_test

import (
	"os"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestConfigOverride(t *testing.T) {
	os.Setenv("CSI_ENDPOINT", "/test-endpoint.sock")
	cfg := config.ProvideConfig(zap.NewNop())
	assert.Equal(t, "/test-endpoint.sock", cfg.Endpoint)
}
