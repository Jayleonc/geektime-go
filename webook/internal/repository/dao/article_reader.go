package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
}

type ArticleGORMReaderDAO struct {
	db *gorm.DB
}

func (a ArticleGORMReaderDAO) Upsert(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func (a ArticleGORMReaderDAO) UpsertV2(ctx context.Context, art PublishedArticle) error {
	panic("implement me")
}

func NewArticleGORMReaderDAO(db *gorm.DB) *ArticleGORMReaderDAO {
	return &ArticleGORMReaderDAO{
		db: db,
	}
}
