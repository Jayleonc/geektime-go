package validator

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/slice"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events/producer"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	l logger.Logger

	producer  producer.Producer
	direction string
	batchSize int
	utime     int64
	// <= 0 中断
	// > 0 睡眠
	sleepInterval time.Duration

	fromBase func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.Logger,
	p producer.Producer) *Validator[T] {
	res := &Validator[T]{base: base, target: target,
		l: l, producer: p, direction: direction, batchSize: 100}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var ego errgroup.Group
	ego.Go(func() error {
		return v.ValidateBaseToTarget(ctx)
	})

	ego.Go(func() error {
		return v.ValidateTargetToBase(ctx)
	})
	return ego.Wait()
}

func (v *Validator[T]) ValidateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		var src T
		src, err := v.fromBase(ctx, offset)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 增量较量，考虑一直运行，不能 return
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}

		if err != nil {
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			offset++
			continue
		}

		// 查到数据后
		var dst T
		err = v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case err == nil:
			equal := src.CompareTo(dst)
			if !equal {
				// 将消息放到 kafka
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 记录日志，然后继续，做好监控
			v.l.Error("base -> target 查询 target 失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
		offset++
	}
}
func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
		Order("utime").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) ValidateBaseToTargetForBatch(ctx context.Context) error {
	offset := 0 // 初始化偏移量为0
	for {
		var srcTs []T
		// 从base数据库中批量获取数据
		//err := v.base.WithContext(ctx).Order("id").Offset(offset).Limit(v.batchSize).Find(&srcTs).Error
		err := v.base.WithContext(ctx).Where("utime > ?", v.utime).
			Order("utime").Offset(offset).Limit(v.batchSize).Find(&srcTs).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || len(srcTs) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			offset += len(srcTs)
			continue
		}

		// 构建ID列表，用于查询目标数据库
		ids := slice.Map(srcTs, func(idx int, t T) int64 {
			return t.ID()
		})

		var dstTs []T
		// 根据ID列表，批量查询目标数据库
		err = v.target.WithContext(ctx).Where("id IN ?", ids).Find(&dstTs).Error
		if err != nil {
			v.l.Error("base -> target 查询 target 失败", logger.Error(err))
			offset += len(srcTs)
			continue
		}

		// 构建目标数据的映射，方便后续比对
		dstMap := make(map[int64]T)
		for _, dst := range dstTs {
			dstMap[dst.ID()] = dst
		}

		// 遍历源数据，检查和比对
		for _, src := range srcTs {
			dst, exists := dstMap[src.ID()]
			if !exists {
				// 目标数据库中缺失数据
				v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
			} else {
				// 比对数据是否一致
				equal := src.CompareTo(dst)
				if !equal {
					// 数据不一致
					v.notify(src.ID(), events.InconsistentEventTypeNEQ)
				}
			}
		}

		if len(srcTs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(srcTs) // 更新偏移量，准备获取下一批数据
	}
}

func (v *Validator[T]) ValidateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		err := v.target.WithContext(ctx).
			Select("id").
			Order("id").
			Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 target 失败", logger.Error(err))
			offset += len(ts)
			continue
		}

		var srcTs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").Where("id IN ?", ids).Find(&srcTs).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || len(ts) == 0 {
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 base 失败", logger.Error(err))
			offset += len(ts)
			continue
		}
		diffs := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		// diffs 里的就是 target 有，base 没有的数据
		v.notifyBaseMissing(diffs)
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

func (v *Validator[T]) notifyBaseMissing(diffs []T) {
	for _, diff := range diffs {
		v.notify(diff.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProducerInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	if err != nil {
		v.l.Error("发送不一致消息失败",
			logger.Error(err),
			logger.String("type", typ),
			logger.Int64("id", id))
	}
}
