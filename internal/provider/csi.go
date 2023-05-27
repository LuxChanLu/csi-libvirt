package provider

import (
	"context"
	"os"
	"strings"

	"github.com/LuxChanLu/csi-libvirt/internal/provider/config"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/controller"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/identity"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/csi/node"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/driver"
	"github.com/LuxChanLu/csi-libvirt/internal/provider/hypervisor"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/exec"
	"k8s.io/utils/mount"
)

func ProvideCSIIdentity(driver *driver.Driver) csi.IdentityServer {
	return &identity.Identity{Driver: driver}
}

func ProvideCSIController(driver *driver.Driver, logger *zap.Logger, hypervisors *hypervisor.Hypervisors) csi.ControllerServer {
	return &controller.Controller{Driver: driver, Logger: logger.With(zap.String("mode", "controller")), Hypervisors: hypervisors}
}

func ProvideCSINode(driver *driver.Driver, logger *zap.Logger, config *config.Config, k8s *kubernetes.Clientset) csi.NodeServer {
	machineIdData, err := os.ReadFile(config.Node.MachineIDFile)
	if err != nil {
		logger.Fatal("unable to read machine id file", zap.String("file", config.Node.MachineIDFile), zap.Error(err))
	}
	machineId := strings.TrimSpace(string(machineIdData))
	node := &node.Node{
		Driver: driver, Logger: logger.With(zap.String("mode", "node"), zap.String("machine-id", machineId)),
		MachineID: machineId,
		Formatter: &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: exec.New()},
		Mounter:   mount.New(""),
	}
	if config.Zone.NodeLabel != "" {
		k8sNode, err := k8s.CoreV1().Nodes().Get(context.Background(), config.Node.Name, v1.GetOptions{})
		if err != nil {
			logger.Fatal("unable to get node information", zap.Error(err))
		}
		zone, ok := k8sNode.Labels[config.Zone.NodeLabel]
		if ok {
			logger.Info("node in a zone detected", zap.String("zone", zone))
			node.Zone = zone
		} else {
			logger.Warn("unable to find node label", zap.String("zone-label", config.Zone.NodeLabel))
		}
	}
	return node
}
