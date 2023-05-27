package driver

import (
	"embed"
	"sync"
	"text/template"

	"github.com/LuxChanLu/csi-libvirt/internal"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/hypervisor"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

//go:embed template/*.tpl
var templates embed.FS

type Driver struct {
	Name     string
	Version  string
	Endpoint string

	Hypervisors *hypervisor.Hypervisors
	Logger      *zap.Logger

	tpl       *template.Template
	logger    *zap.Logger
	diskLocks *sync.Map
}

func ProvideControllerDriver(config *config.Config, hypervisors *hypervisor.Hypervisors, log *zap.Logger) *Driver {
	driver := ProvideNodeDriver(config, log)
	driver.Hypervisors = hypervisors
	return driver
}

func ProvideNodeDriver(config *config.Config, log *zap.Logger) *Driver {
	tpl, err := template.ParseFS(templates, "template/*.tpl")
	if err != nil {
		log.Fatal("unable to parse driver template", zap.Error(err))
	}
	return &Driver{
		Name:     config.DriverName,
		Version:  internal.BuildVersion,
		Endpoint: config.Endpoint,

		Logger: log.With(zap.String("tier", "driver")),

		tpl:       tpl,
		logger:    log,
		diskLocks: &sync.Map{},
	}
}

func RegisterController(srv *grpc.Server, identity csi.IdentityServer, controller csi.ControllerServer) {
	csi.RegisterIdentityServer(srv, identity)
	csi.RegisterControllerServer(srv, controller)
}

func RegisterNode(srv *grpc.Server, identity csi.IdentityServer, node csi.NodeServer) {
	csi.RegisterIdentityServer(srv, identity)
	csi.RegisterNodeServer(srv, node)
}
