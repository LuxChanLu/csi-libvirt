package config

import (
	"os"

	"github.com/traefik/paerser/env"
	"go.uber.org/zap"
)

type Config struct {
	DriverName string
	Endpoint   string

	Node *ConfigNode
}

type ConfigNode struct {
	MachineIDFile string
	Endpoint      string
}

const (
	driverName = "lu.lxc.csi.libvirt"
)

func ProvideConfig(logger *zap.Logger) *Config {
	config := &Config{
		DriverName: driverName,
		Endpoint:   "/var/lib/kubelet/plugins/" + driverName + "/csi.sock",
		Node: &ConfigNode{
			MachineIDFile: "/etc/machine-id",
		},
	}
	if err := env.Decode(os.Environ(), "CSI_", config); err != nil {
		logger.Fatal("unable to parse config from args", zap.Error(err))
	}
	logger.Info("config loaded", zap.String("driverName", config.DriverName))
	return config
}
