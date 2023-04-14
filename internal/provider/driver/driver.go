package driver

import (
	"embed"
	"text/template"

	"github.com/LuxChanLu/libvirt-csi/internal"
	"github.com/LuxChanLu/libvirt-csi/internal/provider/config"
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

	tpl    *template.Template
	logger *zap.Logger
}

func ProvideDriver(config *config.Config, log *zap.Logger) *Driver {
	tpl, err := template.ParseFS(templates, "template/*.tpl")
	if err != nil {
		log.Fatal("unable to parse driver template", zap.Error(err))
	}
	return &Driver{
		Name:     config.DriverName,
		Version:  internal.BuildVersion,
		Endpoint: config.Endpoint,

		tpl:    tpl,
		logger: log,
	}
}

func StartController(srv *grpc.Server, identity csi.IdentityServer, controller csi.ControllerServer) {
	csi.RegisterIdentityServer(srv, identity)
	csi.RegisterControllerServer(srv, controller)
}

func StartNode(srv *grpc.Server, identity csi.IdentityServer, node csi.NodeServer) {
	csi.RegisterIdentityServer(srv, identity)
	csi.RegisterNodeServer(srv, node)
}
