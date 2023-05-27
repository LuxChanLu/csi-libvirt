//go:build !integration

package provider_test

import (
	"os"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/node"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProvideCSIIdentity(t *testing.T) {
	assert.NotNil(t, provider.ProvideCSIIdentity(&driver.Driver{Endpoint: "/tmp/csi-libvirt-test.sock"}))
}

func TestProvideCSIController(t *testing.T) {
	assert.NotNil(t, provider.ProvideCSIController(&driver.Driver{Endpoint: "/tmp/csi-libvirt-test.sock"}, zap.NewNop(), nil))
}

func TestProvideCSINode(t *testing.T) {
	machineId := uuid.New().String()
	assert.NoError(t, os.WriteFile("/tmp/machine-id", []byte(machineId), 0660))
	defer func() {
		assert.NoError(t, os.Remove("/tmp/machine-id"))
	}()
	node := provider.ProvideCSINode(&driver.Driver{Endpoint: "/tmp/csi-libvirt-test.sock"}, zap.NewNop(), &config.Config{Node: &config.ConfigNode{MachineIDFile: "/tmp/machine-id"}}, nil).(*node.Node)
	assert.NotNil(t, node)
	assert.Equal(t, machineId, node.MachineID)
}
