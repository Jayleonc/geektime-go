//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"github.com/jayleonc/geektime-go/webook/interactive/grpc"
	"github.com/jayleonc/geektime-go/webook/interactive/repository"
	"github.com/jayleonc/geektime-go/webook/interactive/repository/cache"
	"github.com/jayleonc/geektime-go/webook/interactive/repository/dao"
	"github.com/jayleonc/geektime-go/webook/interactive/service"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitSaramaClient,
	InitSyncProducer,
	InitLogger,
)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
