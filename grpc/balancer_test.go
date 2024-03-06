package grpc

import (
	"context"
	"fmt"
	_ "github.com/jayleonc/geektime-go/webook/pkg/grpcx/balancer/wrr"
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

type BalancerServer struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *BalancerServer) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:2379")
	require.NoError(s.T(), err)
	s.cli = cli
}

func (s *BalancerServer) TestServer() {
	go func() {
		s.startBalancerServer(":8090", 89, &Server{
			Name: ":8090",
		})
	}()
	//go func() {
	//	s.startBalancerServer(":8091", 83, &Server{
	//		Name: ":8091",
	//	})
	//}()
	//go func() {
	//	s.startBalancerServer(":8092", 67, &Server{
	//		Name: ":8092",
	//	})
	//}()
	s.startBalancerServer(":8093", 75, &FailedServer{
		Name: ":8093",
	})
}

func (s *BalancerServer) startBalancerServer(addr string, weight int, svc UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr = "127.0.0.1" + addr
	key := "service/user/" + addr

	l, err := net.Listen("tcp", addr)
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
				Addr: addr,
				Metadata: map[string]any{
					"weight": weight,
					"now":    now.String(),
				},
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
	RegisterUserServiceServer(server, svc)
	err = server.Serve(l)
	if err != nil {
		panic(err)
	}

	// 取消续约
	kCancel()

}

func (s *BalancerServer) TestFailoverClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"round_robin": {}}],
  "methodConfig":  [
    {
      "name": [{"service":  "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerServer) TestClientCustomWRR() {
	t := s.T()

	builder, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)

	cc, err := grpc.Dial(
		"etcd:///service/user",
		grpc.WithResolvers(builder),
		grpc.WithDefaultServiceConfig(`{ "loadBalancingConfig": [ { "custom_weighted_round_robin": {} } ] }`),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 表示不需要 TLS
	)
	require.NoError(t, err)

	client := NewUserServiceClient(cc)

	fromMap := make(map[string]int)
	for i := 0; i < 100; i++ {
		res, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
		require.NoError(t, err)
		//t.Log(res.User)
		fromMap[res.User.Name]++
	}

	fmt.Println("===============================================")
	fmt.Println(fromMap)
}

func (s *BalancerServer) TestClientWRR() {
	t := s.T()

	builder, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)

	cc, err := grpc.Dial(
		"etcd:///service/user",
		grpc.WithResolvers(builder),
		grpc.WithDefaultServiceConfig(`{ "loadBalancingConfig": [ { "weighted_round_robin": {} } ] }`),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 表示不需要 TLS
	)
	require.NoError(t, err)

	client := NewUserServiceClient(cc)

	fromMap := make(map[string]int)
	for i := 0; i < 1000; i++ {
		res, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
		require.NoError(t, err)
		//t.Log(res.User)
		fromMap[res.User.Name]++
	}

	fmt.Println("===============================================")
	fmt.Println(fromMap)
}

func (s *BalancerServer) TestClient() {
	t := s.T()

	builder, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)

	cc, err := grpc.Dial(
		"etcd:///service/user",
		grpc.WithResolvers(builder),
		grpc.WithDefaultServiceConfig(`{ "loadBalancingConfig": [ { "round_robin": {} } ] }`),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 表示不需要 TLS
	)
	require.NoError(t, err)

	client := NewUserServiceClient(cc)

	fromMap := make(map[string]int)
	for i := 0; i < 3; i++ {
		res, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
		require.NoError(t, err)
		//t.Log(res.User)
		fromMap[res.User.Name]++
	}

	fmt.Println("===============================================")
	fmt.Println(fromMap)
}

func TestRun(t *testing.T) {
	suite.Run(t, new(BalancerServer))
}
