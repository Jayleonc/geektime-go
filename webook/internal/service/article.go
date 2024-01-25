package service

import (
	"context"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	events "github.com/jayleonc/geektime-go/webook/internal/events/article"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	GetByAuthor(ctx context.Context, uid int64, pageIndex, pageSize int, title string) ([]domain.Article, int64, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer events.Producer
	l        logger.Logger
}

func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := a.repo.GetPubById(ctx, id)
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
	return art, nil
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, pageIndex int, pageSize int, title string) ([]domain.Article, int64, error) {
	return a.repo.GetByAuthor(ctx, uid, pageIndex, pageSize)
}

func (a *articleService) Sync(ctx context.Context, art domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func NewArticleService(repo repository.ArticleRepository, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	if article.Id > 0 {
		err := a.repo.Update(ctx, article)
		return article.Id, err
	}
	return a.repo.Create(ctx, article)
}

func (a *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, article)
}
