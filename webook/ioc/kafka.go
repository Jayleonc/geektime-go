package ioc

import (
	"github.com/IBM/sarama"
	"github.com/jayleonc/geektime-go/webook/internal/events"
	"github.com/jayleonc/geektime-go/webook/internal/events/article"
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

// RegisterConsumers 注册 Consumer
func RegisterConsumers(c *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c}
}
