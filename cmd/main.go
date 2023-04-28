package main

import (
	"github.com/LuxChanLu/csi-libvirt/internal/options"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/traefik/paerser/cli"
	"go.uber.org/fx"
)

func main() {
	cmd := &cli.Command{
		Name:        "csi-libvirt",
		Description: "LibVirt CSI",
	}
	err := cmd.AddCommand(&cli.Command{
		Name:        "controller",
		Description: "Controller Server",
		Run: func(s []string) error {
			fx.New(options.AppOptions(fx.Provide(driver.ProvideControllerDriver), fx.Invoke(driver.RegisterController))...).Run()
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
			fx.New(options.AppOptions(fx.Provide(driver.ProvideNodeDriver), fx.Invoke(driver.RegisterNode))...).Run()
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
