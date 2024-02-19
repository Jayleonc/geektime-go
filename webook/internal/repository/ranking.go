package repository

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, articles []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	cache cache.RankingCache

	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

func NewCachedRankingRepositoryV1(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) *CachedRankingRepository {
	return &CachedRankingRepository{redisCache: redisCache, localCache: localCache}
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}

func (c *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := c.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = c.redisCache.Get(ctx)
	if err != nil {
		return c.localCache.ForceGet(ctx)
	}
	_ = c.localCache.Set(ctx, res)
	return res, nil
}

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func (c *CachedRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.cache.Set(ctx, arts)
}