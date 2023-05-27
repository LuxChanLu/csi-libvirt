package provider

import (
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ProvideK8S(log *zap.Logger) *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("failed to get in-cluster config", zap.Error(err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("failed to create clientset", zap.Error(err))
	}
	return clientset
}
