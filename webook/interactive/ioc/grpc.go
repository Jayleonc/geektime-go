package ioc

import (
	grpc2 "github.com/jayleonc/geektime-go/webook/interactive/grpc"
	"github.com/jayleonc/geektime-go/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	s := grpc.NewServer()
	intrSvc.Register(s)
	return &grpcx.Server{
		Server: s,
		Addr:   viper.GetString("grpc.server.addr"),
	}
}
