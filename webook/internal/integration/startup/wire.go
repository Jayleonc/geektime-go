//go:build wireinject

package startup

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

var ThirdPartySet = wire.NewSet(
	InitRedis, InitDB,
	InitLogger, InitWeChatService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ThirdPartySet,
		// DAO 部分
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		// cache 部分
		cache.NewCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
		web.NewArticleHandler,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		ThirdPartySet,
		//dao.NewArticleGORMDAO,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
