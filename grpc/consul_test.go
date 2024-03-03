package grpc

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/jayleonc/geektime-go/webook/pkg/netx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"strconv"
	"testing"
	"time"
)

/**
1.使用 webook/docker-compose.yaml 启动一个 consul 容器
2.先调用 TestConsulServer 向注册中心注册一个服务，可访问 http://localhost:8500/ui/dc1/services 查看服务健康状态
3.调用 TestConsulClient 发现服务，并返回 user.GetByID() 结果
*/

const (
	consulPort  = "8500"
	serviceName = "user_service"
	serviceID   = "user_service_1"
	servicePort = 50051
)

func TestConsulServer(t *testing.T) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
	assert.NoError(t, err)

	err = registerServiceWithConsul()
	assert.NoError(t, err)

	s := grpc.NewServer()
	RegisterUserServiceServer(s, &Server{})

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	// 设置服务状态为SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	t.Logf("Server listening at %v", lis.Addr())
	err = s.Serve(lis)
	assert.NoError(t, err)
	s.GracefulStop() // grpc 优雅退出
}

func TestConsulClient(t *testing.T) {
	serviceAddress, err := discoverServiceFromConsul(serviceName)
	t.Logf("Service addr is %v", serviceAddress)
	assert.NoError(t, err)

	conn, err := grpc.Dial(serviceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	user := NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := user.GetByID(ctx, &GetByIDRequest{Id: 123})
	assert.NoError(t, err)
	t.Logf("Name is : %s", r.User.Name)
}

func registerServiceWithConsul() error {
	config := api.DefaultConfig()
	config.Address = netx.GetOutboundIP() + ":" + consulPort
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: netx.GetOutboundIP(),
		Tags:    []string{"hello"},
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%v:%v", netx.GetOutboundIP(), servicePort), // 注意修改为正确的格式
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		return err
	}
	return nil
}
func discoverServiceFromConsul(serviceName string) (string, error) {
	config := api.DefaultConfig()
	config.Address = netx.GetOutboundIP() + ":" + consulPort
	client, err := api.NewClient(config)
	if err != nil {
		return "", err
	}

	services, err := client.Agent().Services()
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("service %s was not found", serviceName)
	}

	var addr string
	for _, v := range services {
		addr = v.Address + ":" + strconv.Itoa(v.Port)
	}
	return addr, nil
}
