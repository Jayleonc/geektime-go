package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/interactive/repository/dao"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	"github.com/jayleonc/geektime-go/webook/pkg/gormx/connpool"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events/fixer"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events/producer"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/scheduler"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

// InitGinxServer 管理后台的 server
func InitGinxServer(l logger.Logger,
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer producer.Producer) *ginx.Server {
	engine := gin.Default()
	group := engine.Group("/migrator")
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	sch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) producer.Producer {
	return producer.NewSaramaProducer("inconsistent_interactive", p)
}

func InitFixerConsumer(client sarama.Client,
	l logger.Logger,
	src SrcDB,
	dst DstDB) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
