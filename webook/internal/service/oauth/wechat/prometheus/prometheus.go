package prometheus

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/service/oauth/wechat"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Decorator struct {
	wechat.Service
	sum prometheus.Summary
}

func NewWechatDecorator(service wechat.Service, sum prometheus.Summary) *Decorator {
	prometheus.MustRegister(sum)
	return &Decorator{Service: service, sum: sum}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
