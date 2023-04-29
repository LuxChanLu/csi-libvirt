//go:build integration

package identity_test

import (
	"context"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/identity"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestGetPluginInfo(t *testing.T) {
	identity := &identity.Identity{Driver: &driver.Driver{Name: "NameTest", Version: "VersionTest"}}
	pluginInfo, err := identity.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	assert.NoError(t, err)
	assert.Equal(t, identity.Driver.Name, pluginInfo.Name)
	assert.Equal(t, identity.Driver.Version, pluginInfo.VendorVersion)
}

func TestGetPluginCapabilities(t *testing.T) {
	identity := &identity.Identity{Driver: &driver.Driver{Name: "NameTest", Version: "VersionTest"}}
	capabilities, err := identity.GetPluginCapabilities(context.Background(), &csi.GetPluginCapabilitiesRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, capabilities.Capabilities)
}

func TestProbe(t *testing.T) {
	logger := zap.NewNop()
	config := config.ProvideConfig(logger)
	lc := fxtest.NewLifecycle(t)
	lv := provider.ProvideLibvirt(lc, logger, config)
	lc.RequireStart()
	defer lc.RequireStop()

	result, err := (&identity.Identity{Driver: &driver.Driver{Name: "NameTest", Version: "VersionTest"}, Libvirt: lv}).Probe(context.Background(), &csi.ProbeRequest{})
	assert.NoError(t, err)
	assert.True(t, result.Ready.Value)
	result, err = (&identity.Identity{Libvirt: libvirt.NewWithDialer(dialers.NewLocal(dialers.WithSocket("unix:///tmp/not-existing.sock")))}).Probe(context.Background(), &csi.ProbeRequest{})
	assert.NoError(t, err)
	assert.False(t, result.Ready.Value)

	result, err := (&identity.Identity{Driver: &driver.Driver{Name: "NameTest", Version: "VersionTest"}}).Probe(context.Background(), &csi.ProbeRequest{})
	assert.NoError(t, err)
	assert.True(t, result.Ready.Value)
}
