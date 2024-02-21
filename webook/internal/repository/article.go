package repository

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"gorm.io/gorm"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, biz string, article domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]domain.Article, int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
	GetByIds(ctx context.Context, ids []int64) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache

	userRepo UserRepository

	db *gorm.DB
}

func (c *CachedArticleRepository) GetByIds(ctx context.Context, ids []int64) ([]domain.Article, error) {
	// 尝试从缓存中批量获取文章数据
	articles, missingIds, err := c.cache.GetArticlesByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	var missingArticles []domain.Article

	// 如果有缓存未命中的ID，从数据库获取这些文章数据
	if len(missingIds) > 0 {
		missingArticlesDao, err := c.dao.GetByIds(ctx, missingIds)
		if err != nil {
			return nil, err
		}

		// 转换 dao.PublishedArticle 到 domain.Article
		for _, daoArt := range missingArticlesDao {
			domainArt := c.toDomain(dao.Article(daoArt))
			missingArticles = append(missingArticles, domainArt)
		}

		// 将从数据库中获取的文章添加到最终结果列表中
		articles = append(articles, missingArticles...)

		// 异步更新缓存
		go func() {
			if er := c.cache.SetArticles(ctx, missingArticles); er != nil {
				fmt.Println(er)
			}
		}()
	}

	return articles, nil
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	arts, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](arts,
		func(idx int, src dao.PublishedArticle) domain.Article {
			return c.toDomain(dao.Article(src))
		}), nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 查询对应的 User 信息, 拿到创作者信息
	res = c.toDomain(dao.Article(art))
	user, err := c.userRepo.FindById(ctx, res.Author.Id)
	if err != nil {
		// 记录日志
		return domain.Article{}, err
	}
	res.Author.Name = user.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	article := c.toDomain(art)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.Set(ctx, article)
		if er != nil {
			// 记录日志
		}
	}()

	return article, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, pageIndex int, pageSize int) ([]domain.Article, int64, error) {
	// 判断要不要查缓存
	offset := pageIndex - 1
	if offset == 0 && pageSize == 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err != nil {
			return res, int64(len(res)), err
		}
	}

	articles, count, err := c.dao.GetByAuthor(ctx, uid, pageIndex, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var res []domain.Article
	for _, art := range articles {
		d := c.toDomain(art)
		res = append(res, d)
	}

	go func() {
		if offset == 0 && pageSize == 100 {
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志 or do something
			}
		}
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, count, nil
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err != nil {
		return 0, err
	}
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	// 在这里尝试，设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 你可以灵活设置过期时间
		user, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			// 要记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()
	return id, nil
}

// SyncV2 两个DAO,表示有两个数据库,作者的制作库和读者的线上库
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// 防止后面业务panic
	defer tx.Rollback()

	authorDAO := dao.NewArticleGORMAuthorDAO(tx)
	readerDAO := dao.NewArticleGORMReaderDAO(tx)

	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = authorDAO.Update(ctx, artn)
	} else {
		id, err = authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil

}

func NewCachedArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, userRepo UserRepository) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(article))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, article.Author.Id)
		if er != nil {
			return 0, er
		}
		return id, nil
	}
	return 0, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, biz string, article domain.Article) error {
	// 首先更新数据库中的文章
	err := c.dao.UpdateById(ctx, c.toEntity(article))
	if err != nil {
		return err
	}
	// update 需要判断是否存在与 topN 存在就要把article内容更新到 topN 相关的 hash 中
	// 检查文章是否在Top N列表中
	isInTopN, err := c.cache.IsArticleInTopN(ctx, biz, article.Id)
	if err != nil {
		// 处理错误，可能记录日志或返回
		return err
	}

	// 如果文章在Top N列表中，更新缓存中的文章内容
	if isInTopN {
		err = c.cache.UpdateArticleInCache(ctx, biz, article)
		if err != nil {
			// 处理错误，可能记录日志或返回
			return err
		}
	}
	return nil
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录缓存
		}
	}
}
