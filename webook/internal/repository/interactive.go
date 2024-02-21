package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// BatchIncrReadCnt bizs 和 bizIds 长度必须一致
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
	GetTopNLikedArticles(ctx context.Context, biz string, n int) ([]domain.ArticleLike, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) GetTopNLikedArticles(ctx context.Context, biz string, n int) ([]domain.ArticleLike, error) {
	// 尝试从缓存获取数据
	inters, err := c.cache.GetTopNLikedInteractive(ctx, biz, n)
	if err == nil && len(inters) >= n { // 如果缓存命中且数据量充足
		fmt.Println("命中缓存，得到有序集合")
		return inters, nil
	}

	// 通常，N 不会经常变动。
	// 如果缓存未命中或数据不足，从 DB 获取点赞数前 N 的 Interactive 数据
	interactives, err := c.dao.GetTopNLikedInteractive(ctx, biz, n)
	if err != nil {
		return nil, err
	}

	var articleLikes []domain.ArticleLike
	// 转换 interactives 切片到 ArticleLike 结构
	for _, interactive := range interactives {
		articleLike := domain.ArticleLike{
			ArticleId: interactive.BizId,
			LikeCnt:   interactive.LikeCnt,
		}
		articleLikes = append(articleLikes, articleLike)
	}

	// 异步更新缓存以反映最新的Top N数据
	// 下次进来的时候，就能从缓存中直接获取
	go func() {
		if er := c.cache.SetTopNLikedInteractive(ctx, biz, articleLikes); er != nil {
			// 在实际应用中，这里应该记录日志或进行其他错误处理
		}
	}()

	return articleLikes, nil
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(intrs, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	// 更新 DB 的点赞数
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	// 尝试更新已存在的文章点赞数
	if err := c.cache.IncrLikeCntIfPresent(ctx, biz, id); err != nil {
		// 处理错误，可能记录日志等
	}

	// 无条件更新Top N列表的点赞数
	if err := c.cache.IncrLikeCnt(ctx, biz, id, 1); err != nil {
		return err
	}

	return nil
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	if err := c.cache.DecrLikeCntIfPresent(ctx, biz, id); err != nil {
		return err
	}
	// 无条件更新Top N列表的点赞数（减1）
	return c.cache.DecrLikeCnt(ctx, biz, id, -1)
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return intr, nil
	}

	ie, err := c.dao.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}
	if err == nil {
		res := c.toDomain(ie)
		err = c.cache.Set(ctx, biz, id, res)
		if err != nil {
			// 记录日志
			// 回写缓存失败
		}
		return res, err
	}
	return intr, err
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}
func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, cache: cache}
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 更新缓存
	// 如果缓存更新失败, 数据不一致, 无所谓, 用户和作者都感知不到哈哈
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	if err != nil {
		return err
	}
	go func() {
		for i := 0; i < len(bizs); i++ {
			err = c.cache.IncrReadCntIfPresent(ctx, bizs[i], bizIds[i])
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	return nil
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}
