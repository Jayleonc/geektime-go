package service

import (
	"github.com/google/uuid"
	"github.com/jayleonc/geektime-go/webook/internal/domain/async"
)

// NewRetryTask 创建 RetryTask 并为任务生成唯一ID
func NewRetryTask(taskName, taskType, parameters string, count int) async.Task {
	return async.Task{
		Id:         generateUniqueID(),
		Name:       taskName,
		Type:       taskType,
		Parameters: parameters,
		RetryCount: count,
	}
}

// 生成唯一ID
func generateUniqueID() string {
	return uuid.New().String()
}
