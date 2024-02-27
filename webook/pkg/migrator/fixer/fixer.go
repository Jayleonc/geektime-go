package fixer

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/slice"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	rows, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	return &OverrideFixer[T]{base: base, target: target, columns: columns}, err
}

func NewOverrideFixerV1[T migrator.Entity](base *gorm.DB, target *gorm.DB,
	columns []string) *OverrideFixer[T] {
	return &OverrideFixer[T]{base: base, target: target, columns: columns}
}

func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var t T
	err := f.base.WithContext(ctx).Where("id = ?", id).First(&t).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// 没找到直接去 target 删除，因为 base 数据库没有，说明 target 是多的数据
		return f.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case err == nil:
		// upsert，在 base 找到数据，直接覆盖
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	default:
		return err
	}
}

func (f *OverrideFixer[T]) FixForBatch(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// 开始一个事务
	tx := f.target.WithContext(ctx).Begin()
	defer tx.Rollback() // 最终回滚，如果是 Commit 了，回滚也不会出错，即回滚一个已经提交的事务，不会出错。
	if tx.Error != nil {
		// 开启事务失败，还有什么好说的？
		// 如果tx.Begin()失败，通常意味着有更严重的数据库连接问题，可能需要特别注意或记录
		return tx.Error
	}

	var srcTs []T
	err := f.base.WithContext(ctx).Where("id IN ?", ids).Find(&srcTs).Error
	if err != nil {
		return err
	}

	// 获取查询到的所有ID
	var foundIds []int64
	foundIds = slice.Map(srcTs, func(idx int, src T) int64 { return src.ID() })
	idsToDel := slice.DiffSet(ids, foundIds)

	// 在 ids 中但不在 foundIds 中的 ID
	if len(idsToDel) > 0 {
		// 如果有需要删除的记录
		if err := tx.Where("id IN ?", idsToDel).Delete(new(T)).Error; err != nil {
			return err
		}
	}

	// 准备批量修复到 target 数据库中
	// 使用 CreateInBatches 进行批量 upsert 操作
	batchSize := 200
	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns(f.columns),
	}).CreateInBatches(&srcTs, batchSize).Error

	if err != nil {
		return err
	}

	// 成功提交事务，如果提交发生错误，直接返回，defer 会回滚
	return tx.Commit().Error
}

func (f *OverrideFixer[T]) FixV1(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNEQ, events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.Where("id = ?", evt.ID).First(&t).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return f.target.Model(&t).Delete("id = ?", evt.ID).Error
		case err == nil:
			// upsert
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Model(new(T)).Delete("id = ?", evt.ID).Error
	}
	return nil
}
