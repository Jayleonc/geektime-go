package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/web/middleware"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	initViper()
	initLogger()
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello 启动成功啦")
	})

	server.Run(":8080")
}

func initViper() {
	c := pflag.String("config", "config/config.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*c)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func useSessions(engine *gin.Engine) {
	// 初始化 session，并设置校验中间件
	login := &middleware.LoginMiddlewareBuilder{}
	//store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("#D@2A8Kun$zKkuCvyAbkUUdNAWiGuvXf"), []byte("BJEGHveod69QGCd^3FzEwxHFMBrD$3nJ"))
	//store, err := redis.NewStore(8, "tcp", "175.178.58.198:16379", "jayleonc", []byte("#D@2AdKun$zKkuCvyAbkUUdNAWiGuvXf"), []byte("BJEGHveod69QGCd^3FzEwxHFMBrD$3nJ"))
	//if err != nil {
	//	panic(err)
	//}
	engine.Use(sessions.Sessions("ssid", store), login.CheckLogin())
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://localhost:2379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}
