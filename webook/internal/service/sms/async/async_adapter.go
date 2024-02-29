package async

import "context"

type ServiceAdapter struct {
	AsyncService *SmsService
}

func (a *ServiceAdapter) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 调用 asyncService 的 Send 方法
	return a.AsyncService.Send(ctx, tpl, args, numbers...)
}

func (a *ServiceAdapter) Execute() error {
	// 调用 asyncService 的 Execute 方法
	return a.AsyncService.Execute()
}
