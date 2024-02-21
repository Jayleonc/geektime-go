package dao

import (
	"context"
	"errors"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, entity Article) error
	Sync(ctx context.Context, entity Article) (int64, error)
	GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]Article, int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
	GetByIds(ctx context.Context, ids []int64) ([]PublishedArticle, error)
}

type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content  string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
}

type PublishedArticle Article

type ArticleGORMDAO struct {
	db *gorm.DB
}

func (a *ArticleGORMDAO) GetByIds(ctx context.Context, ids []int64) ([]PublishedArticle, error) {
	var articles []PublishedArticle
	if len(ids) == 0 {
		return articles, nil
	}

	// 使用 GORM 的 Where 方法构建查询，以 `id IN (?)` 条件查找已发表的文章
	if err := a.db.WithContext(ctx).
		Where("id IN ? AND status = ?", ids, domain.ArticleStatusPublished).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	return articles, nil
}

func (a *ArticleGORMDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()
	var res []PublishedArticle
	err := a.db.WithContext(ctx).
		Where("utime < ? AND status = ?",
			start.UnixMilli(), domain.ArticleStatusPublished).
		Offset(offset).Limit(limit).
		Find(&res).Error
	return res, err
}

func (a *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := a.db.WithContext(ctx).
		Where("id = ?", id).
		First(&res).Error
	return res, err
}

func (a *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).
		Where("id = ?", id).First(&art).Error
	return art, err
}

func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? and author_id = ?", uid, id).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		return tx.Model(&PublishedArticle{}).
			Where("id = ?", uid).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})
}

func (a *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64, pageIndex int, pageSize int) ([]Article, int64, error) {
	var arts []Article
	var count int64
	if err := a.db.WithContext(ctx).Count(&count).Where("author_id = ?", uid).
		Offset((pageIndex - 1) * pageSize).Limit(pageSize).Order("utime desc").
		Find(&arts).Error; err != nil {
		return nil, 0, err
	}
	return arts, count, nil
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

// Sync The same library has different tables, the author's production library (table) and the reader's online library (table), using different structures to represent different tables, Article and PublishedArticle
func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)
		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle(art) // 另一张表
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   now,
				"status":  pubArt.Status,
			}),
		}).Create(&pubArt).Error
		return err
	})
	return id, err
}

func (a *ArticleGORMDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	err := a.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", art.Id).
		Where("author_id = ?", art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("发生错误，Id或作者有误")
	}
	return nil
}
