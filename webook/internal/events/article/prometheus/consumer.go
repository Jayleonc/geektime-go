package prometheus

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/jayleonc/geektime-go/webook/internal/events/article"
	"github.com/jayleonc/geektime-go/webook/pkg/saramax"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type InteractiveReadEventConsumerWithMetrics struct {
	*article.InteractiveReadEventConsumer
	consumeDurationVec *prometheus.SummaryVec
	consumeCounterVec  *prometheus.CounterVec
}

func NewInteractiveReadEventConsumerWithMetrics(consumer *article.InteractiveReadEventConsumer) *InteractiveReadEventConsumerWithMetrics {
	consumeDurationVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "kafka_consumer_consume_duration",
		Help:      "统计 Kafka 消费者消费操作的执行时间",
	}, []string{"topic", "status"})

	consumeCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "kafka_consumer_consume_count",
		Help:      "统计 Kafka 消费者消费操作的计数",
	}, []string{"topic", "status"})
	prometheus.MustRegister(consumeDurationVec, consumeCounterVec)

	return &InteractiveReadEventConsumerWithMetrics{
		InteractiveReadEventConsumer: consumer,
		consumeDurationVec:           consumeDurationVec,
		consumeCounterVec:            consumeCounterVec,
	}
}

func (r *InteractiveReadEventConsumerWithMetrics) Consume(msg *sarama.ConsumerMessage, t article.ReadEvent) error {
	startTime := time.Now()

	err := r.InteractiveReadEventConsumer.Consume(msg, t)

	duration := time.Since(startTime).Seconds()
	status := "success"
	if err != nil {
		status = "failure"
	}
	r.consumeDurationVec.WithLabelValues(msg.Topic, status).Observe(duration)
	r.consumeCounterVec.WithLabelValues(msg.Topic, status).Inc()

	return err
}

func (r *InteractiveReadEventConsumerWithMetrics) Start() error {
	cgroup, err := sarama.NewConsumerGroupFromClient("interactive", r.GetClient())
	if err != nil {
		return err
	}

	go func() {
		er := cgroup.Consume(context.Background(), []string{article.ReadEventTopic}, saramax.NewHandler[article.ReadEvent](r.Consume))
		if er != nil {
			fmt.Println("退出消费循环", er)
		}
	}()

	return err
}
