//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/jayleonc/geektime-go/webook/interactive/events"
	"github.com/jayleonc/geektime-go/webook/interactive/events/prometheus"
	"github.com/jayleonc/geektime-go/webook/interactive/grpc"
	"github.com/jayleonc/geektime-go/webook/interactive/ioc"
	"github.com/jayleonc/geektime-go/webook/interactive/repository"
	"github.com/jayleonc/geektime-go/webook/interactive/repository/cache"
	"github.com/jayleonc/geektime-go/webook/interactive/repository/dao"
	"github.com/jayleonc/geektime-go/webook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitSrcDB,
	ioc.InitDstDB,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.NewSyncProducer,
	ioc.InitRedis,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(thirdPartySet,
		interactiveSvcSet,
		grpc.NewInteractiveServiceServer,

		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		ioc.InitInteractiveProducer,
		ioc.InitFixerConsumer,
		ioc.InitGinxServer,

		events.NewInteractiveReadEventConsumer,
		prometheus.NewInteractiveReadEventConsumerWithMetrics,
		wire.Struct(new(App), "*"),
	)

	return new(App)
}
