package grpcx

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/jayleonc/geektime-go/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	EtcdAddr string
	Name     string
	Port     int
	L        logger.Logger

	client  *etcdv3.Client
	kCancel func() //  一个取消函数，用于停止维持etcd租约的KeepAlive过程
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	err = s.register() // 在etcd中注册服务
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.client = client

	em, err := endpoints.NewManager(client, "service/"+s.Name)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var ttl int64 = 5
	grant, err := client.Grant(ctx, ttl)

	if err != nil {
		return err
	}
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{Addr: addr}, etcdv3.WithLease(grant.ID))

	if err != nil {
		return err
	}

	kCtx, kCancel := context.WithCancel(context.Background())
	s.kCancel = kCancel
	ch, err := client.KeepAlive(kCtx, grant.ID)
	go func() {
		for c := range ch {
			s.L.Debug(c.String())
		}
	}()
	return err
}

func (s *Server) Close() error {
	if s.kCancel != nil {
		s.kCancel()
	}

	if s.client != nil {
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}
