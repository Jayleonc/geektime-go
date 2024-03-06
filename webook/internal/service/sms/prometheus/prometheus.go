package prometheus

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"time"
)

// Decorator 增加对发送 SMS 操作性能监控的功能
type Decorator struct {
	svc sms.Service
	// 用于收集和记录操作的性能指标。在这里，它会记录发送 SMS 的操作时长
	// SummaryVec 是一种特殊形式的 Summary，允许收集带有一组标签的 Summary 指标
	vector *prometheus.SummaryVec
}

func NewSMSDecorator(svc sms.Service, opts prometheus.SummaryOpts) sms.Service {
	// 用于记录不同 tpl_id 下的操作时长
	vec := prometheus.NewSummaryVec(opts, []string{"tpl_id"})
	// 将 vec 注册到 Prometheus 的默认注册表中，以便 Prometheus 可以抓取指标
	prometheus.MustRegister(vec)
	return &Decorator{svc: svc, vector: vec}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		// 将时长 duration 作为观察值添加到 Summary 向量 vector 中，使用对应的 tpl_id 作为标签
		d.vector.WithLabelValues(tplId).Observe(float64(duration))
	}()
	n := rand.Int63n(99)
	i := rand.Int63n(9)
	time.Sleep(time.Millisecond*time.Duration(n) + time.Millisecond*time.Duration(i))
	return d.svc.Send(ctx, tplId, args, numbers...)
}
