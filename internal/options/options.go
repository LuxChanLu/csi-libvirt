package options

import (
	"github.com/LuxChanLu/csi-libvirt/internal"
	"github.com/LuxChanLu/csi-libvirt/internal/provider"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/hypervisor"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func AppOptions(opts ...fx.Option) []fx.Option {
	return append([]fx.Option{
		fx.Provide(provider.ProvideLogger),
		fx.Provide(fx.Annotate(hypervisor.ProvideLibvirtDialer, fx.ResultTags(`group:"libvirt.dialers"`))),
		fx.Provide(hypervisor.ProvideLibvirt),
		fx.Provide(config.ProvideConfig),
		fx.Provide(provider.ProvideGRPCServer),
		fx.Provide(provider.ProvideCSIIdentity),
		fx.Provide(provider.ProvideCSIController),
		fx.Provide(provider.ProvideCSINode),
		fx.Provide(provider.ProvideK8S),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			logger.With(zap.String("buildTime", internal.BuildTime), zap.String("buildCommit", internal.BuildCommit), zap.String("buildVersion", internal.BuildVersion)).Info("init driver")
			return &fxevent.ZapLogger{Logger: logger}
		}),
	}, opts...)
}
