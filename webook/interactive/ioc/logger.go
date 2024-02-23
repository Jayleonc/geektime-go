package ioc

import (
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitLogger() logger.Logger {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	build, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(build)
}
