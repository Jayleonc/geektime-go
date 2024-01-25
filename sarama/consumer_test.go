package sarama

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type ConsumerHandler struct {
}

func (c ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("This is Setup")
	partitions := session.Claims()["test_topic"]
	//var offset int64 = 0
	for _, partition := range partitions {
		session.ResetOffset("test_topic", partition, sarama.OffsetOldest, "")
	}
	return nil
}

func (c ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("This is Cleanup")
	return nil
}

func (c ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messages := claim.Messages()
	for message := range messages {
		fmt.Println(string(message.Value))
		session.MarkMessage(message, "")
	}
	return nil
}

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup([]string{"175.178.58.198:9094"}, "demo", cfg)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = consumer.Consume(ctx, []string{"test_topic"}, ConsumerHandler{})
	assert.NoError(t, err)
}
