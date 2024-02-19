package command

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/cmd/wire"
	"github.com/jayleonc/geektime-go/webook/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func NewWebookCommand() *cobra.Command {
	w := &cobra.Command{
		Use:   "web",
		Short: "webook web server start.",
		Run:   runWebookWebServer,
	}
	w.PersistentFlags().StringVarP(&Flags.Config, "config", "c", "config/config.yaml", "config file")
	return w
}

func runWebookWebServer(cmd *cobra.Command, args []string) {
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
	initPrometheus()

	otel := ioc.InitOTEL()
	defer func() {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		otel(ctx)
	}()

	app := wire.InitWebServer()
	// 启动 Kafka 消费者
	for _, consumer := range app.Consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}
	// 启动定时任务
	app.Corn.Start()
	defer func() {
		// 等待定时任务退出
		<-app.Corn.Stop().Done()
	}()
	// 启动 Web
	server := app.Web
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello 启动成功啦")
	})

	server.Run(":8080")
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
