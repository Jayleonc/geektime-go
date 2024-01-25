package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleAuthorDAO interface {
	Create(ctx context.Context, article Article) (int64, error)
	Update(ctx context.Context, article Article) error
}

type ArticleGORMAuthorDAO struct {
	db *gorm.DB
}

func (a ArticleGORMAuthorDAO) Create(ctx context.Context, article Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (a ArticleGORMAuthorDAO) Update(ctx context.Context, article Article) error {
	//TODO implement me
	panic("implement me")
}

func NewArticleGORMAuthorDAO(db *gorm.DB) *ArticleGORMAuthorDAO {
	return &ArticleGORMAuthorDAO{
		db: db,
	}
}
