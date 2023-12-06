//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB, ioc.InitLogger,
		// DAO 部分
		dao.NewUserDAO,

		// cache 部分
		cache.NewCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWeChatService,
		service.NewUserService,
		service.NewCodeService,

		// handler 部分
		ijwt.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
