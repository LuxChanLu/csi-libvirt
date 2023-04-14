package provider

import (
	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ProvideLibvirt(lc fx.Lifecycle, log *zap.Logger) *libvirt.Libvirt {
	virt := libvirt.NewWithDialer(dialers.NewLocal())
	lc.Append(fx.StartStopHook(virt.Connect, virt.Disconnect))
	return virt
}
