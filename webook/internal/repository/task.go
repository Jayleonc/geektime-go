package repository

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/domain/async"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
)

type AsyncTaskRepository interface {
	StoreTask(ctx context.Context, task async.Task) error
	LoadTasks(ctx context.Context, name string) ([]async.Task, error)
	UpdateTask(ctx context.Context, task async.Task) error
}

type asyncTaskRepository struct {
	dao dao.TaskDAO
}

func (a *asyncTaskRepository) UpdateTask(ctx context.Context, task async.Task) error {
	return a.dao.UpdateTask(ctx, task)
}

func NewAsyncTaskRepository(dao dao.TaskDAO) AsyncTaskRepository {
	return &asyncTaskRepository{dao: dao}
}

func (a *asyncTaskRepository) LoadTasks(ctx context.Context, name string) ([]async.Task, error) {
	tasks, err := a.dao.LoadTask(ctx, name)
	if err != nil {
		return nil, err
	}
	var aTasks []async.Task
	for _, t := range tasks {
		aTasks = append(aTasks, a.toDomain(t))
	}
	return aTasks, nil
}

func (a *asyncTaskRepository) StoreTask(ctx context.Context, task async.Task) error {
	entity := a.toEntity(task)
	return a.dao.StoreTask(ctx, entity)
}

func (a *asyncTaskRepository) toEntity(task async.Task) dao.Task {
	return dao.Task{
		Id:           task.Id,
		Name:         task.Name,
		Type:         task.Type,
		Parameters:   task.Parameters,
		Status:       dao.TaskStatus(task.Status),
		ErrorMessage: task.ErrorMessage,
		RetryCount:   task.RetryCount,
	}
}

func (a *asyncTaskRepository) toDomain(task dao.Task) async.Task {
	return async.Task{
		Id:           task.Id,
		Name:         task.Name,
		Type:         task.Type,
		Parameters:   task.Parameters,
		Status:       int(task.Status),
		ErrorMessage: task.ErrorMessage,
		RetryCount:   task.RetryCount,
	}
}
