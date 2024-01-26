package prometheus

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/events/article"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type KafkaProducerWithMetrics struct {
	article.Producer
	durationVec *prometheus.SummaryVec
	counterVec  *prometheus.CounterVec
}

func NewKafkaProducerWithMetrics(producer article.Producer) article.Producer {
	durationVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "kafka_producer_send_duration",
		Help:      "统计 Kafka 生产者发送操作执行时间",
	}, []string{"topic", "status"})

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "kafka_producer_send_count",
		Help:      "统计 Kafka 生产者发送操作的计数",
	}, []string{"topic", "status"})
	prometheus.MustRegister(durationVec, counterVec)
	return &KafkaProducerWithMetrics{Producer: producer, durationVec: durationVec, counterVec: counterVec}
}

func (k *KafkaProducerWithMetrics) ProduceReadEvent(ctx context.Context, evt article.ReadEvent) error {
	startTime := time.Now()
	err := k.Producer.ProduceReadEvent(ctx, evt)
	status := "success"
	if err != nil {
		status = "failure"
	}
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		k.durationVec.WithLabelValues(article.ReadEventTopic, status).Observe(float64(duration))
		k.counterVec.WithLabelValues(article.ReadEventTopic, status).Inc()
	}()

	return err
}
