package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error
	GetTopNLikedInteractive(ctx context.Context, biz string, n int) ([]domain.ArticleLike, error)
	SetTopNLikedInteractive(ctx context.Context, biz string, likes []domain.ArticleLike) error
	IncrLikeCnt(ctx context.Context, biz string, id int64, i int) error
	DecrLikeCnt(ctx context.Context, biz string, id int64, decrement int64) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) DecrLikeCnt(ctx context.Context, biz string, id int64, decrement int64) error {
	// 构造Top N列表的key
	topNKey := "topn:likes:" + biz
	// 将文章ID转换为字符串作为member
	member := strconv.FormatInt(id, 10)

	// 使用ZINCRBY命令减少点赞数
	return i.client.ZIncrBy(ctx, topNKey, float64(decrement), member).Err()
}

func (i *InteractiveRedisCache) IncrLikeCnt(ctx context.Context, biz string, id int64, increment int) error {
	// 构造 TopN 列表的key
	topNKey := "topn:likes:" + biz
	// 将文章 ID 转换为字符串作为 member
	member := strconv.FormatInt(id, 10)

	// 使用 ZINCRBY 命令更新点赞数
	return i.client.ZIncrBy(ctx, topNKey, float64(increment), member).Err()
}

func (i *InteractiveRedisCache) SetTopNLikedInteractive(ctx context.Context, biz string, likes []domain.ArticleLike) error {
	// 使用 Redis 的 pipeline 优化性能
	pipe := i.client.Pipeline()

	key := "topn:likes:" + biz
	for _, al := range likes {
		// 更新有序集合中的点赞数
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(al.LikeCnt),
			Member: al.ArticleId,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetTopNLikedInteractive 获取缓存中点赞数最高的 Top N 篇文章的点赞数和文章ID
func (i *InteractiveRedisCache) GetTopNLikedInteractive(ctx context.Context, biz string, n int) ([]domain.ArticleLike, error) {
	key := "topn:likes:" + biz

	// 获取点赞数最高的N篇文章的ID及其点赞数
	results, err := i.client.ZRevRangeWithScores(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	var topArticles []domain.ArticleLike
	for _, result := range results {
		id, _ := strconv.ParseInt(result.Member.(string), 10, 64)
		likeCnt := int64(result.Score)
		topArticles = append(topArticles, domain.ArticleLike{
			ArticleId: id,
			LikeCnt:   likeCnt,
		})
	}

	return topArticles, nil
}

func (i *InteractiveRedisCache) Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error {
	key := i.key(biz, bizId)
	err := i.client.HSet(ctx, key, fieldReadCnt, res.ReadCnt, fieldLikeCnt, res.LikeCnt, fieldCollectCnt, res.CollectCnt).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	key := i.key(biz, id)
	res, err := i.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	var intr domain.Interactive
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	return intr, nil
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Err()
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, 1).Err()
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Err()
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{client: client}
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Err()
}

func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
