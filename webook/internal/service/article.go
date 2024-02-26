package service

import (
	"context"
	"fmt"
	intrv1 "github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	events "github.com/jayleonc/geektime-go/webook/internal/events/article"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, biz string, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	GetByAuthor(ctx context.Context, uid int64, pageIndex, pageSize int, title string) ([]domain.Article, int64, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetByIds(ctx context.Context, ids []int64) ([]domain.Article, error)
	GetPubById(ctx context.Context, biz string, id int64, uid int64) (domain.Article, *intrv1.GetResponse, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)

	Like(ctx context.Context, biz string, id, uid int64, like bool) error
	Collect(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	GetTopNArticles(ctx context.Context, biz string, n int) ([]domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer events.Producer
	l        logger.Logger
}

func (a *articleService) GetTopNArticles(ctx context.Context, biz string, n int) ([]domain.Article, error) {

	var sortedArticles []domain.Article
	req := &intrv1.GetTopNLikedArticlesRequest{
		Biz: biz,
		N:   int32(n),
	}
	topIntrArticles, err := a.repo.GetTopNIntr(ctx, req)
	if err != nil {
		return nil, err
	}

	// 构建文章ID的切片
	ids := make([]int64, len(topIntrArticles.ArticleLike))
	for i, al := range topIntrArticles.ArticleLike {
		ids[i] = al.ArticleId
	}

	// 获取的是 TopN 的文章ID、标题和摘要
	articles, err := a.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	articlesMap := make(map[int64]domain.Article)
	for _, article := range articles {
		articlesMap[article.Id] = article
	}

	for _, id := range ids {
		if article, exists := articlesMap[id]; exists {
			sortedArticles = append(sortedArticles, article)
		}
	}
	return sortedArticles, nil
}

func (a *articleService) Collect(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	return a.repo.Collect(ctx, biz, id, cid, uid)
}

func (a *articleService) Like(ctx context.Context, biz string, id, uid int64, like bool) error {
	return a.repo.Like(ctx, biz, id, uid, like)
}

func (a *articleService) GetByIds(ctx context.Context, ids []int64) ([]domain.Article, error) {
	return a.repo.GetByIds(ctx, ids)
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

func (a *articleService) GetPubById(ctx context.Context, biz string, id, uid int64) (domain.Article, *intrv1.GetResponse, error) {
	art, intr, err := a.repo.GetPubById(ctx, biz, id, uid)
	// 添加阅读计数
	if err == nil {
		go func() {
			er := a.producer.ProduceReadEvent(
				ctx,
				events.ReadEvent{
					Aid: id,
					Uid: uid,
				},
			)
			if er != nil {
				fmt.Println("发送读者阅读事件失败", err)
			}
		}()
	}
	return art, intr, nil
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, pageIndex int, pageSize int, title string) ([]domain.Article, int64, error) {
	return a.repo.GetByAuthor(ctx, uid, pageIndex, pageSize)
}

func NewArticleService(repo repository.ArticleRepository, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) Save(ctx context.Context, biz string, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	if article.Id > 0 {
		err := a.repo.Update(ctx, biz, article)
		return article.Id, err
	}
	return a.repo.Create(ctx, article)
}

func (a *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, article)
}
