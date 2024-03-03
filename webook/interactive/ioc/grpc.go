package ioc

import (
	grpc2 "github.com/jayleonc/geektime-go/webook/interactive/grpc"
	"github.com/jayleonc/geektime-go/webook/pkg/grpcx"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer, l logger.Logger) *grpcx.Server {
	type Config struct {
		EtcdAddr string `yaml:"etcdAddr"`
		Name     string `yaml:"name"`
		Port     int    `yaml:"port"`
	}

	s := grpc.NewServer()
	intrSvc.Register(s)
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	return &grpcx.Server{
		Server:   s,
		EtcdAddr: cfg.EtcdAddr,
		Name:     cfg.Name,
		Port:     cfg.Port,
		L:        l,
	}
}
