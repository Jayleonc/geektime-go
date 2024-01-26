package prometheus

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Decorator struct {
	svc sms.Service
	//svc    localsms.Service
	vector *prometheus.SummaryVec
}

func NewSMSDecorator(svc sms.Service, opts prometheus.SummaryOpts) sms.Service {
	vec := prometheus.NewSummaryVec(opts, []string{"tpl_id"})
	prometheus.MustRegister(vec)
	return &Decorator{svc: svc, vector: vec}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.vector.WithLabelValues(tplId).Observe(float64(duration))
	}()
	return d.svc.Send(ctx, tplId, args, numbers...)
}
