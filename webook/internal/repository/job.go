package repository

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"time"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, id int64, time time.Time) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, time time.Time) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, jid int64) error {
	return p.dao.UpdateUtime(ctx, jid)
}

func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	//TODO implement me
	panic("implement me")
}

func NewPreemptJobRepository(dao dao.JobDAO) CronJobRepository {
	return &PreemptJobRepository{dao: dao}
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}
