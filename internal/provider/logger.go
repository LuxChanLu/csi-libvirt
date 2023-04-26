package provider

import (
	"go.uber.org/zap"
)

func ProvideLogger() *zap.Logger {
	zap, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return zap
}
