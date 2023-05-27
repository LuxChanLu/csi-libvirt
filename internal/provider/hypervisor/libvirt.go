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

type provideLibvirtParams struct {
	fx.In

	Log     *zap.Logger
	Dialers []*ZonedDialer `group:"libvirt.dialers"`
}

func ProvideLibvirt(lc fx.Lifecycle, params *provideLibvirtParams) *Hypervisors {
	virts := make([]*ZonedHypervisor, len(params.Dialers))
	for idx, dialer := range params.Dialers {
		virts[idx] = &ZonedHypervisor{Libvirt: libvirt.NewWithDialer(dialer), Zone: dialer.Zone}
	}
	lc.Append(fx.StartStopHook(func() error {
		for _, virt := range virts {
			if err := virt.Connect(); err != nil {
				return err
			}
			version, err := virt.ConnectGetLibVersion()
			if err != nil {
				return err
			}
			params.Log.Info("libvirt connected", zap.Uint64("version", version))
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
	return &Hypervisors{Libvirts: virts, Logger: params.Log.With(zap.String("tier", "hypervisor"))}
}

func ProvideLibvirtDialer(log *zap.Logger, config *config.Config) []*ZonedDialer {
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
	dialers := make([]*ZonedDialer, len(config.Zone.Zones)+1)
	dialers[0] = buildDialer(config.LibvirtEndpoint)
	for idx, zone := range config.Zone.Zones {
		dialers[idx+1] = buildDialer(zone.LibvirtEndpoint)
		dialers[idx+1].Zone = zone.Name
	}
	return dialers
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
