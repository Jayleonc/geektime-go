package article

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/pkg/saramax"
	"time"
)

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client}
}

// Start 负责初始化 Kafka 消费者组并开始消费消息
func (r *InteractiveReadEventConsumer) Start() error {
	cgroup, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}

	go func() {
		er := cgroup.Consume(context.Background(), []string{"read_article"}, saramax.NewBatchHandler[ReadEvent](r.BatchConsume))
		if er != nil {
			fmt.Println("退出消费循环", er)
		}
	}()

	return err
}

// StartV1 负责初始化 Kafka 消费者组并开始消费消息，单个消费
func (r *InteractiveReadEventConsumer) StartV1() error {
	cgroup, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}

	go func() {
		er := cgroup.Consume(context.Background(), []string{"read_article"}, saramax.NewHandler[ReadEvent](r.Consume))
		if er != nil {
			fmt.Println("退出消费循环", er)
		}
	}()

	return err
}

func (r *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, t []ReadEvent) error {
	bizs := make([]string, len(t))
	bizIds := make([]int64, len(t))
	for _, event := range t {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, event.Aid)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}

func (r *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	fmt.Println("Kafka 执行 incrReadCnt")
	return r.repo.IncrReadCnt(ctx, "article", t.Aid)
}
