package cache

import (
	"context"
	"encoding/json"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, articles []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func NewRankingRedisCache(client redis.Cmdable) RankingCache {
	return &RankingRedisCache{client: client, key: "ranking:top_n", expiration: time.Minute * 3}
}

func (r *RankingRedisCache) Set(ctx context.Context, articles []domain.Article) error {
	for i := range articles {
		articles[i].Content = articles[i].Abstract()
	}

	bytes, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, bytes, r.expiration).Err()
}
