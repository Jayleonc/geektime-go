package producer

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events"
)

type Producer interface {
	ProducerInconsistentEvent(ctx context.Context, evt events.InconsistentEvent) error
}
type SaramaProducer struct {
	p     sarama.SyncProducer
	topic string
}

func NewSaramaProducer(topic string, p sarama.SyncProducer) *SaramaProducer {
	return &SaramaProducer{
		topic: topic,
		p:     p,
	}
}
func (s *SaramaProducer) ProducerInconsistentEvent(ctx context.Context, evt events.InconsistentEvent) error {
	val, _ := json.Marshal(evt)
	_, _, err := s.p.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(val),
	})
	return err
}
