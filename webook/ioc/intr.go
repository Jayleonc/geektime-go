package ioc

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	intrv1 "github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"github.com/jayleonc/geektime-go/webook/interactive/service"
	"github.com/jayleonc/geektime-go/webook/internal/client"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	resolver2 "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewIntrClientV1(client *etcdv3.Client) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}

	resolver, err := resolver2.NewBuilder(client)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithResolvers(resolver)}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	dial, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}

	return intrv1.NewInteractiveServiceClient(dial)
}

func NewIntrClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool
		Threshold int32
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}

	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	dial, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}

	remote := intrv1.NewInteractiveServiceClient(dial)
	local := client.NewLocalInteractiveServiceAdapter(svc)
	res := client.NewInteractiveClient(remote, local)
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置监控，发生配置更改事件")
		cfg = Config{}
		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			panic(err)
		}

		res.UpdateThreshold(cfg.Threshold)
	})

	return res
}
