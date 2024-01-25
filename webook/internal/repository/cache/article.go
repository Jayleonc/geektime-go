package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	DelFirstPage(ctx context.Context, id int64) error
	SetPub(ctx context.Context, art domain.Article) error
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	bytes, err := a.client.Get(ctx, a.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(bytes, &res)
	return res, err
}

func (a *ArticleRedisCache) Set(ctx context.Context, article domain.Article) error {
	val, err := json.Marshal(&article)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.key(article.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	bytes, err := a.client.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(bytes, &res)
	return res, err
}

func NewArticleRedisCache(c redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{
		client: c,
	}
}

func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, id int64) error {
	key := a.firstKey(id)
	if err := a.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (a *ArticleRedisCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(&art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.pubKey(art.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error {
	for i := range res {
		res[i].Content = res[i].Abstract()
	}
	key := a.firstKey(uid)
	val, err := json.Marshal(&res)
	if err != nil {
		return err
	}
	err = a.client.Set(ctx, key, val, time.Minute*10).Err()
	if err != nil {
		return err
	}
	return nil
}

func (a *ArticleRedisCache) firstKey(id int64) string {
	return fmt.Sprintf("article:first_page:%d", id)
}

func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func (a *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}
