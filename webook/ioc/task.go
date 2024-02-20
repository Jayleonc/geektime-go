package ioc

import (
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/async"
	"github.com/jayleonc/geektime-go/webook/pkg/retry"
)

func InitTask(smsTask *async.SmsService, demoTask *service.Demo) *retry.Scheduler {
	scheduler := retry.NewScheduler()
	scheduler.RegisterTask(smsTask, demoTask)
	return scheduler
}
