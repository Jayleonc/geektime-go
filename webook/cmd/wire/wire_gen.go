// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/async"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	"github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, logger)
	db := ioc.InitDB(logger)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	taskDAO := dao.NewTaskDAO(db)
	asyncTaskRepository := repository.NewAsyncTaskRepository(taskDAO)
	smsService := ioc.InitSMSServiceV1(asyncTaskRepository, logger)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := ioc.InitWeChatService()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	articleDAO := dao.NewArticleGORMDAO(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, articleCache, userRepository)
	client := ioc.InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := ioc.NewKafkaProducerWithMetricsDecorator(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDAO := dao2.NewGORMInteractiveDAO(db)
	interactiveCache := cache2.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository2.NewCachedInteractiveRepository(interactiveDAO, interactiveCache)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	interactiveServiceClient := ioc.NewIntrClient(interactiveService)
	articleHandler := web.NewArticleHandler(logger, articleService, interactiveServiceClient)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(interactiveRepository, client)
	interactiveReadEventConsumerWithMetrics := prometheus.NewInteractiveReadEventConsumerWithMetrics(interactiveReadEventConsumer)
	v2 := ioc.RegisterConsumers(interactiveReadEventConsumerWithMetrics)
	rankingCache := cache.NewRankingRedisCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache)
	rankingService := service.NewBatchRankingService(interactiveServiceClient, articleService, rankingRepository)
	rlockClient := ioc.InitRLockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingService, logger, rlockClient)
	cron := ioc.InitJobs(logger, rankingJob)
	asyncSmsService := async.NewSmsService(smsService, asyncTaskRepository, logger)
	demo := service.NewDemo()
	scheduler := ioc.InitTask(asyncSmsService, demo)
	app := &App{
		Web:       engine,
		Consumers: v2,
		Corn:      cron,
		Scheduler: scheduler,
	}
	return app
}

// wire.go:

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO, cache2.NewInteractiveRedisCache, repository2.NewCachedInteractiveRepository, service2.NewInteractiveService)

var rankingSvcSet = wire.NewSet(cache.NewRankingRedisCache, repository.NewCachedRankingRepository, service.NewBatchRankingService)

var smsServiceSet = wire.NewSet(async.NewSmsService, ioc.InitSMSServiceV1)
