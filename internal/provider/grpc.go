package provider

import (
	"net"

	"github.com/LuxChanLu/libvirt-csi/internal/provider/driver"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ProvideGRPCServer(lc fx.Lifecycle, driver *driver.Driver, log *zap.Logger) *grpc.Server {
	srv := grpc.NewServer()
	lc.Append(fx.StartStopHook(func() error {
		listener, err := net.Listen("unix", driver.Endpoint)
		if err != nil {
			return err
		}
		go func() {
			if err := srv.Serve(listener); err != nil && err != grpc.ErrServerStopped {
				log.Fatal("unable to start grpc server", zap.Error(err))
			}
		}()
		return nil
	}, srv.GracefulStop))
	return srv
}
