//go:build !integration

package driver_test

import (
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/hypervisor"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTemplate(t *testing.T) {
	driver := driver.ProvideControllerDriver(&config.Config{}, &hypervisor.Hypervisors{}, zap.NewNop())
	assert.NotNil(t, driver.Template("disk.xml.tpl", map[string]interface{}{}))
}
