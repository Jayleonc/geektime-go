package saramax

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type Handler[T any] struct {
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{fn: fn}
}

func (h Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messages := claim.Messages()
	for msg := range messages {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = h.fn(msg, t)
		if err != nil {
			fmt.Println(err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
