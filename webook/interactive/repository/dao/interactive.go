package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error)
	// GetTopNLikedInteractive 得到点赞数前 N 的 文章Id
	GetTopNLikedInteractive(ctx context.Context, biz string, n int) ([]Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

const MaxAllowedN = 100 // 假设100是业务上可接受的最大值

func (dao *GORMInteractiveDAO) GetTopNLikedInteractive(ctx context.Context, biz string, n int) ([]Interactive, error) {
	if n > MaxAllowedN {
		// 如果n超过最大允许值，可以选择返回错误
		return nil, fmt.Errorf("请求的数量超过了最大允许值：%d", MaxAllowedN)
		// 或者将n限制为最大值，静默处理，不返回错误
		// n = MaxAllowedN
	}

	var interactives []Interactive
	if err := dao.db.WithContext(ctx).
		Where("biz = ?", biz).
		Order("like_cnt desc").
		Limit(n).Find(&interactives).Error; err != nil {
		return nil, err
	}
	return interactives, nil
}

func (dao *GORMInteractiveDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	var res []Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id IN ?", biz, ids).
		Find(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? and biz_id = ?", biz, id).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).
		Where("biz = ? and biz_id = ? and uid = ? and status = ?", biz, id, uid, 1).
		First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).Where("biz = ? and biz_id = ? and uid = ?", biz, id, uid).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&UserCollectionBiz{
			Biz:   biz,
			BizId: id,
			Cid:   cid,
			Uid:   uid,
			Ctime: now,
			Utime: now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        biz,
			BizId:      id,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Biz:    biz,
			BizId:  id,
			Status: 1,
			Utime:  now,
			Ctime:  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` + 1"), // todo: what's this?
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("uid = ? and biz_id = ? and biz = ?", uid, id, biz).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).Where("biz = ? and biz_id = ?", biz, id).Updates(map[string]interface{}{
			"like_cnt": gorm.Expr("`like_cnt` - 1"),
			"utime":    now,
		}).Error
	})
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO { return &GORMInteractiveDAO{db: db} }

func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("`read_cnt` + 1"), // todo: what's this?
			"utime":    now,
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   id,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (dao *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDao := NewGORMInteractiveDAO(tx)
		for i := 0; i < len(bizs); i++ {
			err := txDao.IncrReadCnt(ctx, bizs[i], ids[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz    string `gorm:"type:varchar(32);uniqueIndex:uid_biz_type_id"`
	Status int8
	Ctime  int64
	Utime  int64
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`                   // 也就是 文章的 ID
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"` // 业务，也就是文章 article
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

type UserCollectionBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Cid   int64  `gorm:"index"`
	BizId int64  `gorm:"uniqueIndex:biz_type_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_uid"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_uid"`
	Ctime int64
	Utime int64
}

// Collection 收藏夹
type Collection struct {
	Id   int64  `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"type=varchar(1024)"`
	Uid  int64  `gorm:""`

	Ctime int64
	Utime int64
}
