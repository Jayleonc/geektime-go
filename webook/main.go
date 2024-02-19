package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/cmd"
	"github.com/jayleonc/geektime-go/webook/internal/web/middleware"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	cmd.MustStart()
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
