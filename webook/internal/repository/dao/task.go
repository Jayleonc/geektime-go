package dao

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/domain/async"
	"gorm.io/gorm"
	"time"
)

type TaskDAO interface {
	StoreTask(ctx context.Context, task Task) error
	LoadTask(ctx context.Context, name string) ([]Task, error)
	UpdateTask(ctx context.Context, task async.Task) error
}

type taskDAO struct {
	db *gorm.DB
}

func (t *taskDAO) UpdateTask(ctx context.Context, task async.Task) error {
	now := time.Now().UnixMilli()
	return t.db.WithContext(ctx).Model(&Task{}).
		Where("id = ?", task.Id).
		Updates(map[string]any{
			"utime":         now,
			"status":        task.Status,
			"error_message": task.ErrorMessage,
			"retry_count":   task.RetryCount,
		}).Error
}

func (t *taskDAO) LoadTask(ctx context.Context, name string) ([]Task, error) {
	var task []Task
	if err := t.db.WithContext(ctx).Where("name = ? and status = ? and retry_count > ?", name, StatusPending, 0).Find(&task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (t *taskDAO) StoreTask(ctx context.Context, task Task) error {
	// todo 漏了时间
	return t.db.WithContext(ctx).Create(task).Error
}

func NewTaskDAO(db *gorm.DB) TaskDAO {
	return &taskDAO{db: db}
}

// Task 定义了与数据库交互的Task模型
type Task struct {
	Id           string     `gorm:"column:id;type:varchar(36);primaryKey"`
	Name         string     `gorm:"column:name;type:varchar(255);not null"`
	Type         string     `gorm:"column:type;type:varchar(50);not null"`
	Parameters   string     `gorm:"column:parameters;type:text;not null"`
	RetryCount   int        `gorm:"column:retry_count;type:int;not null"`
	Status       TaskStatus `gorm:"column:status;type:int;not null"`
	ErrorMessage string     `gorm:"column:error_message;type:text"`
	CTime        int64      `gorm:"column:ctime"`
	UTime        int64      `gorm:"column:utime"`
}

type TaskStatus int

const (
	StatusPending    TaskStatus = iota // 默认为0，表示待处理
	StatusProcessing                   // 1，表示正在处理
	StatusSuccess                      // 2，表示处理成功
	StatusFailed                       // 3，表示处理失败
)
