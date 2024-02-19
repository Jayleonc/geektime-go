package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/job"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, l logger.Logger, client *rlock.Client) *job.RankingJob {
	return job.NewRankingJob(svc, l, time.Second*30, client)
}

func InitJobs(l logger.Logger, rjob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "定时任务执行",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 1m", builder.Build(rjob))
	if err != nil {
		panic(err)
	}
	return expr
}
