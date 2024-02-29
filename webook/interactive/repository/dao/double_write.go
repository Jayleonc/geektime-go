package dao

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"gorm.io/gorm"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string] // todo 这个是什么类型，学习一下
	l       logger.Logger
}

func NewDoubleWriteDAO(src, dst *gorm.DB, l logger.Logger) *DoubleWriteDAO {
	return &DoubleWriteDAO{
		src:     NewGORMInteractiveDAO(src),
		dst:     NewGORMInteractiveDAO(dst),
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, id)
	case PatternSrcFirst:
		if err := d.src.IncrReadCnt(ctx, biz, id); err != nil {
			return err
		}
		if err := d.dst.IncrReadCnt(ctx, biz, id); err != nil {
			// 要不要 return？
			// 正常来说，我们认为双写阶段，src 成功了就算业务上成功了
			d.l.Error("双写写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, id)
		if err == nil {
			err1 := d.src.IncrReadCnt(ctx, biz, id)
			if err1 != nil {
				d.l.Error("双写写入 src 失败", logger.Error(err1),
					logger.Int64("biz_id", id),
					logger.String("biz", biz))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, id)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.BatchIncrReadCnt(ctx, bizs, ids)
	case PatternSrcFirst:
		if err := d.src.BatchIncrReadCnt(ctx, bizs, ids); err != nil {
			return err
		}
		if err := d.dst.BatchIncrReadCnt(ctx, bizs, ids); err != nil {
			// 要不要 return？
			// 正常来说，我们认为双写阶段，src 成功了就算业务上成功了
			d.l.Error("双写写入 dst 失败", logger.Error(err))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.BatchIncrReadCnt(ctx, bizs, ids)
		if err == nil {
			err1 := d.src.BatchIncrReadCnt(ctx, bizs, ids)
			if err1 != nil {
				d.l.Error("双写写入 src 失败", logger.Error(err1))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.BatchIncrReadCnt(ctx, bizs, ids)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.InsertLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		if err := d.src.InsertLikeInfo(ctx, biz, id, uid); err != nil {
			return err
		}
		if err := d.dst.InsertLikeInfo(ctx, biz, id, uid); err != nil {
			d.l.Error("双写写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstFirst:
		if err := d.dst.InsertLikeInfo(ctx, biz, id, uid); err != nil {
			return err
		}
		if err := d.src.InsertLikeInfo(ctx, biz, id, uid); err != nil {
			d.l.Error("双写写入 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstOnly:
		return d.dst.InsertLikeInfo(ctx, biz, id, uid)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.DeleteLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		if err := d.src.DeleteLikeInfo(ctx, biz, id, uid); err != nil {
			return err
		}
		if err := d.dst.DeleteLikeInfo(ctx, biz, id, uid); err != nil {
			d.l.Error("双写写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstFirst:
		if err := d.dst.DeleteLikeInfo(ctx, biz, id, uid); err != nil {
			return err
		}
		if err := d.src.DeleteLikeInfo(ctx, biz, id, uid); err != nil {
			d.l.Error("双写写入 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstOnly:
		return d.dst.DeleteLikeInfo(ctx, biz, id, uid)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.InsertCollectionBiz(ctx, biz, id, cid, uid)
	case PatternSrcFirst:
		if err := d.src.InsertCollectionBiz(ctx, biz, id, cid, uid); err != nil {
			return err
		}
		if err := d.dst.InsertCollectionBiz(ctx, biz, id, cid, uid); err != nil {
			d.l.Error("双写写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstFirst:
		if err := d.dst.InsertCollectionBiz(ctx, biz, id, cid, uid); err != nil {
			return err
		}
		if err := d.src.InsertCollectionBiz(ctx, biz, id, cid, uid); err != nil {
			d.l.Error("双写写入 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstOnly:
		return d.dst.InsertCollectionBiz(ctx, biz, id, cid, uid)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.GetLikeInfo(ctx, biz, id, uid)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.GetLikeInfo(ctx, biz, id, uid)
	default:
		return UserLikeBiz{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.GetCollectInfo(ctx, biz, id, uid)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.GetCollectInfo(ctx, biz, id, uid)
	default:
		return UserCollectionBiz{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.Get(ctx, biz, id)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.Get(ctx, biz, id)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.GetByIds(ctx, biz, ids)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.GetByIds(ctx, biz, ids)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetTopNLikedInteractive(ctx context.Context, biz string, n int) ([]Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.GetTopNLikedInteractive(ctx, biz, n)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.GetTopNLikedInteractive(ctx, biz, n)
	default:
		return nil, errUnknownPattern
	}
}

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)
