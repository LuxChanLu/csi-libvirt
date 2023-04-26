package driver_test

import (
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTemplate(t *testing.T) {
	driver := driver.ProvideDriver(&config.Config{}, zap.NewNop())
	assert.NotNil(t, driver.Template("disk.xml.tpl", map[string]interface{}{}))
}
