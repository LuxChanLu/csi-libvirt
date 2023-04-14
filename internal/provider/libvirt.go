package provider

import (
	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ProvideLibvirt(lc fx.Lifecycle, log *zap.Logger) *libvirt.Libvirt {
	virt := libvirt.NewWithDialer(dialers.NewLocal())
	lc.Append(fx.StartStopHook(func() error {
		if err := virt.Connect(); err != nil {
			return err
		}
		version, err := virt.Version()
		if err != nil {
			return err
		}
		log.Info("libvirt connected", zap.String("version", version))
		return nil
	}, virt.Disconnect))
	return virt
}