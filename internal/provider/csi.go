package provider

import (
	"github.com/LuxChanLu/libvirt-csi/internal/provider/csi/controller"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/csi/identity"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/csi/node"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/driver"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/go-libvirt"
)

func ProvideCSIIdentity(driver *driver.Driver) csi.IdentityServer {
	return &identity.Identity{Driver: driver}
}

func ProvideCSIController(driver *driver.Driver, libvirt *libvirt.Libvirt) csi.ControllerServer {
	return &controller.Controller{Driver: driver, Libvirt: libvirt}
}

func ProvideCSINode(driver *driver.Driver) csi.NodeServer {
	return &node.Node{Driver: driver}
}
