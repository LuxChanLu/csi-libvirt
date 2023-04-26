package provider

import (
	"errors"
	"net"
	"os"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ProvideGRPCServer(lc fx.Lifecycle, driver *driver.Driver, log *zap.Logger) *grpc.Server {
	srv := grpc.NewServer()
	lc.Append(fx.StartStopHook(func() error {
		_, err := os.Stat(driver.Endpoint)
		if !errors.Is(err, os.ErrNotExist) {
			os.Remove(driver.Endpoint)
		}
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
