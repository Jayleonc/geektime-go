package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
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
	GetArticlesByIds(ctx context.Context, ids []int64) ([]domain.Article, []int64, error)
	SetArticles(ctx context.Context, articles []domain.Article) error
	IsArticleInTopN(ctx context.Context, biz string, id int64) (bool, error)
	UpdateArticleInCache(ctx context.Context, biz string, article domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func (a *ArticleRedisCache) UpdateArticleInCache(ctx context.Context, biz string, article domain.Article) error {
	// 确保键格式与GetArticlesByIds方法一致
	articleKey := fmt.Sprintf("article:%d", article.Id)
	_, err := a.client.HSet(ctx, articleKey, map[string]interface{}{
		"title":    article.Title,
		"abstract": article.Abstract(),
	}).Result()
	return err
}

func (a *ArticleRedisCache) IsArticleInTopN(ctx context.Context, biz string, id int64) (bool, error) {
	// 这个功能，应该放在 Interactive 中，可以考虑在后续将 interactive 抽取层微服务后，进行修改。
	topNKey := "topn:likes:" + biz
	score, err := a.client.ZScore(ctx, topNKey, strconv.FormatInt(id, 10)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil // 文章不在Top N中
	}
	if err != nil {
		return false, err // 执行出错
	}
	return score != 0, nil // 文章在Top N中，score不为0表示存在
}

// SetArticles 将一批文章的详细信息更新到缓存中
// todo 是否需要设置过期时间？
func (a *ArticleRedisCache) SetArticles(ctx context.Context, articles []domain.Article) error {
	for _, article := range articles {
		key := fmt.Sprintf("article:%d", article.Id)
		fields := map[string]interface{}{
			"title":    article.Title,
			"abstract": article.Abstract(),
		}
		if err := a.client.HMSet(ctx, key, fields).Err(); err != nil {
			// todo 做点什么好？
			return err
		}
	}
	return nil
}

// GetArticlesByIds 从缓存中获取一组文章的详细信息，返回找到的文章列表，以及未找到的文章ID列表
func (a *ArticleRedisCache) GetArticlesByIds(ctx context.Context, ids []int64) ([]domain.Article, []int64, error) {
	var foundArticles []domain.Article
	var missingIds []int64

	for _, id := range ids {
		articleData, err := a.client.HGetAll(ctx, fmt.Sprintf("article:%d", id)).Result()
		if err != nil || len(articleData) == 0 {
			missingIds = append(missingIds, id)
			continue
		}

		article := domain.Article{
			Id:      id,
			Title:   articleData["title"],
			Content: articleData["abstract"],
		}
		foundArticles = append(foundArticles, article)
	}

	fmt.Println("命中缓存，得到文章内容集合")
	return foundArticles, missingIds, nil

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
