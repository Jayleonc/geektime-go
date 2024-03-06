package command

import (
	"fmt"
	"github.com/jayleonc/geektime-go/webook/interactive/cmd/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
)

func NewIntrCommand() *cobra.Command {
	w := &cobra.Command{
		Use:   "intr",
		Short: "webook intr server start.",
		Run:   runWebookIntrServer,
	}
	w.PersistentFlags().StringVarP(&Flags.Config, "config", "c", "config/config.yaml", "config file")
	return w
}

func runWebookIntrServer(cmd *cobra.Command, args []string) {
	runApp()
}

func initConfig() {
	viper.SetConfigType("yaml")
	fmt.Println("configPath:", Flags.Config)
	viper.SetConfigFile(Flags.Config)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func runApp() {
	initConfig()
	initLogger()
	//initPrometheus()

	app := wire.InitApp()
	// 启动 Kafka 消费者
	for _, consumer := range app.Consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}

	go func() {
		err1 := app.AdminServer.Start()
		panic(err1)
	}()

	fmt.Println("intr grpc start...")
	if err := app.Server.Serve(); err != nil {
		panic(err)
	}

}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
