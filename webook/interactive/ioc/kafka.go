package ioc

import (
	"github.com/IBM/sarama"
	prometheus2 "github.com/jayleonc/geektime-go/webook/interactive/events/prometheus"
	"github.com/jayleonc/geektime-go/webook/interactive/repository/dao"
	"github.com/jayleonc/geektime-go/webook/internal/events"
	"github.com/jayleonc/geektime-go/webook/internal/events/article"
	"github.com/jayleonc/geektime-go/webook/internal/events/article/prometheus"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events/fixer"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// InitConsumers 注册 Consumer
func InitConsumers(m *prometheus2.InteractiveReadEventConsumerWithMetrics, fixConsumer *fixer.Consumer[dao.Interactive]) []events.Consumer {
	return []events.Consumer{m, fixConsumer}
}

func NewKafkaProducerWithMetricsDecorator(syncProducer sarama.SyncProducer) article.Producer {
	baseProducer := article.NewKafkaProducer(syncProducer)
	return prometheus.NewKafkaProducerWithMetrics(baseProducer)
}
