//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/jayleonc/geektime-go/webook/interactive/events"
	"github.com/jayleonc/geektime-go/webook/interactive/events/prometheus"
	repository2 "github.com/jayleonc/geektime-go/webook/interactive/repository"
	cache2 "github.com/jayleonc/geektime-go/webook/interactive/repository/cache"
	dao2 "github.com/jayleonc/geektime-go/webook/interactive/repository/dao"
	service2 "github.com/jayleonc/geektime-go/webook/interactive/service"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/async"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(cache.NewRankingRedisCache, repository.NewCachedRankingRepository, service.NewBatchRankingService)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB, ioc.InitLogger,
		ioc.InitKafka, ioc.InitRLockClient,
		ioc.RegisterConsumers,
		ioc.NewSyncProducer,

		//async.NewSmsService,
		// 注册 Task 的方法
		ioc.InitTask,
		repository.NewAsyncTaskRepository,
		// DAO 部分
		dao.NewUserDAO,
		dao.NewTaskDAO,

		// cache 部分
		cache.NewCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,
		cache.NewArticleRedisCache,

		events.NewInteractiveReadEventConsumer,
		prometheus.NewInteractiveReadEventConsumerWithMetrics,
		//article.NewKafkaProducer,
		ioc.NewKafkaProducerWithMetricsDecorator,
		interactiveSvcSet,
		ioc.NewIntrClient,
		rankingSvcSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,

		// Service 部分
		smsServiceSet,
		service.NewDemo,
		//ioc.InitSMSService,
		//ioc.InitAsyncSMSService,
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

var smsServiceSet = wire.NewSet(
	ioc.InitAsyncSMSService,
	// 使用 wire.Bind 来绑定接口和实现
	wire.Bind(new(sms.Service), new(*async.SmsService)),
)
