package provider

import (
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ProvideLibvirt(lc fx.Lifecycle, log *zap.Logger, config *config.Config) *libvirt.Libvirt {
	virt := libvirt.NewWithDialer(dialers.NewLocal(dialers.WithSocket(config.QEMUEndpoint)))
	lc.Append(fx.StartStopHook(func() error {
		if err := virt.Connect(); err != nil {
			return err
		}
		version, err := virt.ConnectGetLibVersion()
		if err != nil {
			return err
		}
		log.Info("libvirt connected", zap.Uint64("version", version))
		return nil
	}, virt.Disconnect))
	return virt
}
