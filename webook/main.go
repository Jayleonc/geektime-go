package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/web/middleware"
	"net/http"
)

func main() {

	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello 启动成功啦")
	})

	server.Run(":8080")
}

//func initUserHdl(db *gorm.DB, redisClint redis.Cmdable, server *gin.Engine, codeSvc *service.codeService) {
//	uc := cache.NewUserCache(redisClint)
//	ud := dao.NewUserDAO(db)
//	ur := repository.NewCachedUserRepository(ud, uc)
//	us := service.NewUserService(ur)
//	handler := web.NewUserHandler(us, codeSvc)
//	handler.RegisterRoutes(server)
//}
//
//func initCodeServer(redisClint redis.Cmdable) *service.codeService {
//	cc := cache.NewCodeCache(redisClint)
//	crepo := repository.NewCodeRepository(cc)
//	return service.NewCodeService(crepo, ioc.InitSMSService())
//}

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
