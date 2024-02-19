package service

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"time"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.Logger
	refreshInterval time.Duration
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

func NewCronJobService(repo repository.CronJobRepository, l logger.Logger) CronJobService {
	return &cronJobService{repo: repo, l: l, refreshInterval: time.Minute}
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}

	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()

	j.CancelFunc = func() {
		ticker.Stop() // 如果不关闭，上面的 协程 就会泄漏
		ctx1, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.repo.Release(ctx1, j.Id)
		if er != nil {
			c.l.Error("释放 job 失败",
				logger.Error(er),
				logger.Int64("jib", j.Id))
		}
	}
	return domain.Job{}, err
}

func (c *cronJobService) refresh(id int64) {
	// 本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err),
			logger.Int64("jid", id))
	}
}
