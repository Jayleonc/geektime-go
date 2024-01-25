package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncProducer(t *testing.T) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	producer, err := sarama.NewAsyncProducer([]string{"175.178.58.198:9094"}, config)
	assert.NoError(t, err)
	msgs := producer.Input()
	msgs <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是一条消息"),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 Metadata",
	}
	select {
	case msg := <-producer.Successes():
		t.Log("发送成功：", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors():
		t.Log("发送失败：", err.Err, err.Msg)
	}
}
