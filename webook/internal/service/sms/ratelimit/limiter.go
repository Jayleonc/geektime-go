package ratelimit

import (
	"context"
	"errors"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/jayleonc/geektime-go/webook/pkg/limiter"
	"log"
)

var errLimited = errors.New("触发限流")

// Send 根据需要包含更详细的错误处理
func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		log.Printf("限流触发，键值: %s\n", r.key)
		ctx = WithAsyncMode(ctx, true)
	}
	return r.svc.Send(ctx, tplId, args, numbers...)
}

var _ sms.Service = &RateLimitSMSService{}

type RateLimitSMSService struct {
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func NewRateLimitSMSService(svc sms.Service,
	l limiter.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}

// WithAsyncMode 创建一个新的context，包含一个标记以确认是否执行异步发送。
func WithAsyncMode(ctx context.Context, async bool) context.Context {
	return context.WithValue(ctx, "asyncMode", async)
}
