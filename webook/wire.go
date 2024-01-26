//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/jayleonc/geektime-go/webook/internal/events/article"
	"github.com/jayleonc/geektime-go/webook/internal/events/article/prometheus"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB, ioc.InitLogger,
		ioc.InitKafka,
		ioc.RegisterConsumers,
		ioc.NewSyncProducer,
		// DAO 部分
		dao.NewUserDAO,

		// cache 部分
		cache.NewCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,
		cache.NewArticleRedisCache,

		article.NewInteractiveReadEventConsumer,
		prometheus.NewInteractiveReadEventConsumerWithMetrics,
		//article.NewKafkaProducer,
		ioc.NewKafkaProducerWithMetricsDecorator,
		interactiveSvcSet,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWeChatService,
		service.NewUserService,
		service.NewCodeService,

		dao.NewArticleGORMDAO,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,

		// handler 部分
		ijwt.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
