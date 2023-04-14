package main

import (
	"github.com/LuxChanLu/libvirt-csi/internal"
	"github.com/LuxChanLu/libvirt-csi/internal/provider"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/config"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/driver"
	"github.com/traefik/paerser/cli"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	options := []fx.Option{
		fx.Provide(provider.ProvideLogger),
		fx.Provide(provider.ProvideLibvirt),
		fx.Provide(config.ProvideConfig),
		fx.Provide(driver.ProvideDriver),
		fx.Provide(provider.ProvideGRPCServer),
		fx.Provide(provider.ProvideCSIIdentity),
		fx.Provide(provider.ProvideCSIController),
		fx.Provide(provider.ProvideCSINode),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			logger.With(zap.String("buildTime", internal.BuildTime), zap.String("buildCommit", internal.BuildCommit), zap.String("buildVersion", internal.BuildVersion)).Info("init driver")
			return &fxevent.ZapLogger{Logger: logger}
		}),
	}
	cmd := &cli.Command{
		Name:        "libvirt-csi",
		Description: "LibVirt CSI",
	}
	err := cmd.AddCommand(&cli.Command{
		Name:        "controller",
		Description: "Controller Server",
		Run: func(s []string) error {
			fx.New(append(options, fx.Invoke(driver.StartController))...).Run()
			return nil
		},
	})
	if err != nil {
		panic(err)
	}
	err = cmd.AddCommand(&cli.Command{
		Name:        "node",
		Description: "Node Server",
		Run: func(s []string) error {
			fx.New(append(options, fx.Invoke(driver.StartNode))...).Run()
			return nil
		},
	})
	if err != nil {
		panic(err)
	}

	if err = cli.Execute(cmd); err != nil {
		panic(err)
	}
}
