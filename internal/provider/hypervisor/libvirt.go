package hypervisor

import (
	"fmt"
	"net/url"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Hypervisors struct {
	Libvirts []*ZonedHypervisor
	Logger   *zap.Logger
}

type ZonedHypervisor struct {
	*libvirt.Libvirt
	Zone string
}

type ZonedDialer struct {
	socket.Dialer
	Zone string
}

func ProvideLibvirt(lc fx.Lifecycle, log *zap.Logger, config *config.Config) *Hypervisors {
	dialers := buildDialers(log, config)
	virts := make([]*ZonedHypervisor, len(dialers))
	for idx, dialer := range dialers {
		virts[idx] = &ZonedHypervisor{Libvirt: libvirt.NewWithDialer(dialer), Zone: dialer.Zone}
	}
	lc.Append(fx.StartStopHook(func() error {
		log.Info("connection to hypervisors", zap.Int("count", len(virts)))
		for _, virt := range virts {
			if err := virt.Connect(); err != nil {
				return err
			}
			version, err := virt.ConnectGetLibVersion()
			if err != nil {
				return err
			}
			log.Info("libvirt connected", zap.Uint64("version", version))
		}
		return nil
	}, func() error {
		for _, virt := range virts {
			if err := virt.Disconnect(); err != nil {
				return err
			}
		}
		return nil
	}))
	return &Hypervisors{Libvirts: virts, Logger: log.With(zap.String("tier", "hypervisor"))}
}

func buildDialers(log *zap.Logger, config *config.Config) []*ZonedDialer {
	buildDialer := func(uri string) *ZonedDialer {
		endpoint, err := url.Parse(uri)
		if err != nil {
			log.Fatal("unable to parse libvirt endpoint", zap.String("endpoint", config.LibvirtEndpoint))
		}
		var dialer socket.Dialer
		switch endpoint.Scheme {
		case "tcp":
			log.Info("connect to a tcp dialer", zap.String("hostname", endpoint.Hostname()), zap.String("port", endpoint.Port()))
			opts := []dialers.RemoteOption{}
			if endpoint.Port() != "" {
				opts = append(opts, dialers.UsePort(endpoint.Port()))
			}
			dialer = dialers.NewRemote(endpoint.Hostname(), opts...)
		case "unix":
			log.Info("connect to a unix dialer", zap.String("endpoint", endpoint.Path))
			dialer = dialers.NewLocal(dialers.WithSocket(endpoint.Path))
		default:
			log.Fatal("unimplemented protocol for libvirt", zap.String("protocol", endpoint.Scheme))
		}
		return &ZonedDialer{Dialer: dialer}
	}
	if len(config.Zone.Zones) > 0 {
		dialers := make([]*ZonedDialer, len(config.Zone.Zones))
		for idx, zone := range config.Zone.Zones {
			dialers[idx] = buildDialer(zone.LibvirtEndpoint)
			dialers[idx].Zone = zone.Name
		}
		if len(config.LibvirtEndpoint) > 0 {
			dialers = append(dialers, buildDialer(config.LibvirtEndpoint))
		}
		return dialers
	}
	return []*ZonedDialer{{Dialer: buildDialer(config.LibvirtEndpoint)}}
}

func (h *Hypervisors) Zone(zone string) (*libvirt.Libvirt, error) {
	for _, zonedLibvirt := range h.Libvirts {
		if zonedLibvirt.Zone == zone {
			return zonedLibvirt.Libvirt, nil
		}
	}
	h.Logger.Error("unable to get a libvirt instance for a zone", zap.String("zone", zone))
	return nil, fmt.Errorf("unable to get libvirt instance for zone %s", zone)
}
