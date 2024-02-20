package failover

import (
	"context"
	"errors"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"log"
	"sync/atomic"
)

type FailOverSMSService struct {
	svcs []sms.Service

	idx uint64 // 当前服务商下标，用于故障转移的起始点
}

func NewFailOverSMSService(svcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: svcs,
	}
}

func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}

// 起始下标轮询
// 并且出错也轮询
func (f *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	startIdx := atomic.LoadUint64(&f.idx) // 使用Load确保读取的安全性
	length := uint64(len(f.svcs))

	for i := uint64(0); i < length; i++ {
		// 使用(startIdx + i) % length计算当前索引，确保循环遍历所有服务
		currentIdx := (startIdx + i) % length
		svc := f.svcs[currentIdx]
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			atomic.CompareAndSwapUint64(&f.idx, startIdx, (startIdx+i+1)%length) // 仅在成功时更新idx
			return nil
		}
		// 特定的错误处理
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}
