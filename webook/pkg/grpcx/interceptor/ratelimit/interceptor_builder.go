package ratelimit

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type Builder struct {
	limiter.Limiter
	key     string
	prefixs []string
}

func NewBuilder(limiter limiter.Limiter, key string, prefixs []string) *Builder {
	return &Builder{Limiter: limiter, key: key, prefixs: prefixs}
}

func (b *Builder) BuilderServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limit, err := b.Limiter.Limit(ctx, b.key)
		if err != nil {
			// 保守或激进
			// 保守的做法时，直接拒绝，认定被限流了，不提供服务，认为 redis 崩了，所有请求限流
			// 激进的做法，就算报错了，放行，激进的做法可能会进一步拖累整个应用或其他系统
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		if limit {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}

		return handler(ctx, req)
	}
}

func (b *Builder) BuilderServerUnaryInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if b.prefixs != nil {
			for _, prefix := range b.prefixs {
				if strings.HasPrefix(info.FullMethod, prefix) {
					limit, err := b.Limiter.Limit(ctx, b.key)
					if err != nil {
						// 保守或激进
						// 保守的做法时，直接拒绝，认定被限流了，不提供服务，认为 redis 崩了，所有请求限流
						// 激进的做法，就算报错了，放行，激进的做法可能会进一步拖累整个应用或其他系统
						return nil, status.Errorf(codes.ResourceExhausted, "限流")
					}
					if limit {
						return nil, status.Errorf(codes.ResourceExhausted, "限流")
					}
				}
			}
		}
		return handler(ctx, req)
	}
}
