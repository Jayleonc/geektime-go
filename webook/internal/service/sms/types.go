package sms

import "context"

// Service 发送短信的抽象
// 为了屏蔽不同供应商之间的区别
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}

const (
	AsyncMode      = "asyncMode"
	SkipAuth       = "skipAuth"
	SkipAsyncCheck = "skipAsyncCheck"
)

// WithAsyncMode 创建一个新的 context，包含一个标记以确认是否执行异步发送。
func WithAsyncMode(ctx context.Context, async bool) context.Context {
	return context.WithValue(ctx, AsyncMode, async)
}

// WithSkipAuth 创建一个新的 context，包含一个标志以跳过 JWT 鉴权。
func WithSkipAuth(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, SkipAuth, skip)
}

// WithSkipAsyncCheck 创建一个新的 context，包含一个标记以跳过异步检查。
func WithSkipAsyncCheck(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, SkipAsyncCheck, skip)
}
