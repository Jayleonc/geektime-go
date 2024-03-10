package ratelimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type BuilderV1 struct {
	client     redis.Cmdable
	keyGenFunc func(ctx *gin.Context) string
	window     time.Duration
	rate       int
}

func NewBuilderV1(client redis.Cmdable, window time.Duration, rate int) *BuilderV1 {
	return &BuilderV1{
		client: client,
		window: window,
		rate:   rate,
		keyGenFunc: func(ctx *gin.Context) string {
			return ctx.ClientIP()
		}}
}

func (b *BuilderV1) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 拿到当前时间戳
		now := time.Now().UnixNano()
		windowStart := fmt.Sprintf("%d", now-b.window.Nanoseconds())
		key := b.keyGenFunc(ctx)
		err := b.client.ZRemRangeByScore(ctx, key, "0", windowStart).Err()
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}

		// 统计窗口内还有多少个请求
		reqs := b.client.ZCount(ctx, key, windowStart, fmt.Sprintf("%d", now))
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}

		if reqs.Val() >= int64(b.rate) {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		err = b.client.ZAddNX(ctx, key, redis.Z{
			Score:  float64(now),
			Member: now,
		}).Err()
		// 打日志
		_ = b.client.Expire(ctx, key, b.window).Err()
		ctx.Next()
	}
}
