package config

import (
	"os"

	"github.com/traefik/paerser/env"
	"go.uber.org/zap"
)

type Config struct {
	DriverName      string
	Endpoint        string
	LibvirtEndpoint string

	Zone *ZoneConfig
	Node *ConfigNode
}

type ConfigNode struct {
	Name          string
	MachineIDFile string
}

type ZoneConfig struct {
	NodeLabel string
	Zones     []*Zone
}

type Zone struct {
	Name            string
	LibvirtEndpoint string
}

const (
	driverName = "lu.lxc.csi.libvirt"
)

func ProvideConfig(logger *zap.Logger) *Config {
	config := &Config{
		DriverName:      driverName,
		Endpoint:        "/var/lib/kubelet/plugins/" + driverName + "/csi.sock",
		LibvirtEndpoint: "unix:///var/run/libvirt/libvirt-sock",
		Node: &ConfigNode{
			MachineIDFile: "/etc/machine-id",
		},
		Zone: &ZoneConfig{
			Zones: []*Zone{},
		},
	}
	if err := env.Decode(os.Environ(), "CSI_", config); err != nil {
		logger.Fatal("unable to parse config from env", zap.Error(err))
	}
	logger.Info("config loaded", zap.String("driverName", config.DriverName))
	return config
}
