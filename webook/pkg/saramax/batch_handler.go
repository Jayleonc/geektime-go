package saramax

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"time"
)

type BatchHandler[T any] struct {
	fn func(msgs []*sarama.ConsumerMessage, t []T) error
}

func NewBatchHandler[T any](fn func(msgs []*sarama.ConsumerMessage, t []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{fn: fn}
}

func (b BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	const batchSize = 10
	msgs := claim.Messages()
	for {
		var batch = make([]*sarama.ConsumerMessage, 0, batchSize)
		var ts = make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done = false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					// 代表消费者被关闭了
					return nil
				}
				batch = append(batch, msg)
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					fmt.Println("反序列化消息体失败", err)
					continue
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		// 凑够了一批，然后你就处理
		err := b.fn(batch, ts)
		if err != nil {
			fmt.Println(err)
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}
