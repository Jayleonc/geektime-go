package cache

import (
	"context"
	"errors"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"sync"
	"time"
)

var key = "ranking:article"

type ArticleCacheItem struct {
	TopN      []domain.Article
	ExpiresAt time.Time
}

type RankingLocalCache struct {
	cache      sync.Map
	expiration time.Duration
}

func (r *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	expiration := time.Now().Add(r.expiration)
	item := ArticleCacheItem{
		TopN:      arts,
		ExpiresAt: expiration,
	}
	r.cache.Store(key, item)

	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	value, ok := r.cache.Load(key)
	if !ok {
		return nil, errors.New("本地缓存失效了")
	}
	item := value.(ArticleCacheItem)
	if item.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return item.TopN, nil
}

// ForceGet 不检查是否过期，直接返回过期的数据
func (r *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	value, ok := r.cache.Load(key)
	if !ok {
		return nil, errors.New("本地缓存失效了")
	}

	item := value.(ArticleCacheItem)
	return item.TopN, nil
}
