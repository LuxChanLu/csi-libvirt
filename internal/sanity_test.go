//go:build sanity

package internal_test

import (
	"os"
	"testing"

	"github.com/LuxChanLu/csi-libvirt/internal/options"
	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
	"go.uber.org/fx/fxtest"
)

func TestSanity(t *testing.T) {
	config := sanity.NewTestConfig()
	config.Address = "/tmp/csi.sock"
	os.Setenv("CSI_ENDPOINT", config.Address)

	app := fxtest.New(t, options.AppOptions()...).RequireStart()
	defer app.RequireStop()

	sanity.Test(t, config)
}
