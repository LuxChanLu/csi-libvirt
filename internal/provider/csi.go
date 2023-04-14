package provider

import (
	"io/ioutil"

	"github.com/LuxChanLu/libvirt-csi/internal/provider/config"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/csi/controller"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/csi/identity"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/csi/node"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
	"go.uber.org/zap"
)

func ProvideCSIIdentity(driver *driver.Driver) csi.IdentityServer {
	return &identity.Identity{Driver: driver}
}

func ProvideCSIController(driver *driver.Driver, libvirt *libvirt.Libvirt) csi.ControllerServer {
	return &controller.Controller{Driver: driver, Libvirt: libvirt}
}

func ProvideCSINode(driver *driver.Driver, logger *zap.Logger, config *config.Config) csi.NodeServer {
	machineIdData, err := ioutil.ReadFile(config.Node.MachineIDFile)
	if err != nil {
		logger.Fatal("unable to read machine id file", zap.String("file", config.Node.MachineIDFile), zap.Error(err))
	}
	return &node.Node{Driver: driver, Logger: logger.With(zap.String("mode", "node"), zap.String("machine-id", string(machineIdData))), MachineID: string(machineIdData)}
}
