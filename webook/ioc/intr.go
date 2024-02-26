package ioc

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	intrv1 "github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"github.com/jayleonc/geektime-go/webook/interactive/service"
	"github.com/jayleonc/geektime-go/webook/internal/client"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
