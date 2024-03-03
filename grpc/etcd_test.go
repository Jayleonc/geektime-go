package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:2379")
	// etcdv3.NewFromURLs()
	// etcdv3.New(etcdv3.Config{Endpoints: })
	require.NoError(s.T(), err)
	s.cli = cli
}

func (s *EtcdTestSuite) TestClient() {
	t := s.T()

	builder, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)

	cc, err := grpc.Dial(
		"etcd:///service/user",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := NewUserServiceClient(cc)
	res, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(res.User)
}

func (s *EtcdTestSuite) TestServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr

	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var ttl int64 = 5
	grant, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{Addr: addr}, etcdv3.WithLease(grant.ID))
	require.NoError(t, err)

	go func() {
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			c1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err1 := em.AddEndpoint(c1, key, endpoints.Endpoint{
				Addr:     addr,
				Metadata: now.String(),
			}, etcdv3.WithLease(grant.ID))
			if err1 != nil {
				t.Log(err1)
			}
			cancel1()
		}
	}()

	kCtx, kCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.cli.KeepAlive(kCtx, grant.ID)
		require.NoError(t, err1)
		for response := range ch {
			t.Log(response.String())
		}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	err = server.Serve(l)
	if err != nil {
		panic(err)
	}

	// 取消续约
	kCancel()

	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	server.GracefulStop() // grpc 优雅退出
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
