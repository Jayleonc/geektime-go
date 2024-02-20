package retry

import (
	"context"
	"fmt"
	"time"
)

type AsyncRetry interface {
	Execute() error
}

// Scheduler 负责调度和执行任务
type Scheduler struct {
	tasks  []AsyncRetry
	ctx    context.Context
	cancel func()
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		ctx:    ctx,
		cancel: cancel,
	}
	go func() {
		s.Start()
	}()
	return s
}

// RegisterTask 用于注册任务到调度器
func (s *Scheduler) RegisterTask(task ...AsyncRetry) {
	fmt.Println("注册任务到调度器")
	s.tasks = append(s.tasks, task...)
}

// Start 启动调度器，遍历并执行每个任务
func (s *Scheduler) Start() {
	for _, task := range s.tasks {
		go func(t AsyncRetry) {
			// 定期执行或单次执行取决于任务逻辑
			for {
				if err := t.Execute(); err != nil {
					fmt.Println("任务执行出错:", err)
				}
				time.Sleep(1 * time.Second) // 根据需要调整执行间隔
			}
		}(task)
	}
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cancel()
}
