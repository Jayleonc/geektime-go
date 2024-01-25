package repository

import (
	"context"
	"database/sql"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"log"
	"time"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	FindByWechat(ctx context.Context, id string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (u *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return u.dao.Insert(ctx, u.toEntity(user))
}

func (u *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return u.toDomain(user), nil
}

func (u *CachedUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := u.cache.Get(ctx, uid)
	// 只要 err 为 nil，就返回
	if err == nil {
		return du, nil
	}

	// err 不为 nil，就要查询数据库
	// err 有两种可能
	// 1. key 不存在，说明 redis 是正常的
	// 2. 访问 redis 有问题。可能是网络有问题，也可能是 redis 本身就崩溃了

	user, err := u.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = u.toDomain(user)

	err = u.cache.Set(ctx, du)
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		log.Println(err)
	}
	return du, nil
}

func (u *CachedUserRepository) FindByIdV1(ctx context.Context, uid int64) (domain.User, error) {
	du, err := u.cache.Get(ctx, uid)

	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist: // 缓存没有数据，但 Redis 运作正常
		user, err := u.dao.FindById(ctx, uid)
		if err != nil {
			return domain.User{}, err
		}
		du = u.toDomain(user)
		err = u.cache.Set(ctx, du)
		return du, nil
	default: // redis 不正常，不去查数据库，类似于降级，避免大量请求打到数据库
		return domain.User{}, err
	}
}

func (u *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := u.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return u.toDomain(user), nil
}

func (u *CachedUserRepository) UpdateNonZeroFields(ctx context.Context,
	user domain.User) error {
	// 更新 DB 之后，删除
	err := u.dao.UpdateById(ctx, u.toEntity(user))
	if err != nil {
		return err
	}
	// 延迟一秒
	//time.AfterFunc(time.Second, func() {
	//	_ = u.cache.Del(ctx, user.Id)
	//})
	//return u.cache.Del(ctx, user.Id)
	return nil
}

func (u *CachedUserRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Phone:    user.Phone.String,
		Password: user.Password,
		Nickname: user.Nickname,
		AboutMe:  user.AboutMe,
		Birthday: time.UnixMilli(user.Birthday),
		Ctime:    time.UnixMilli(user.Ctime),
		Utime:    time.UnixMilli(user.Utime),
		WechatInfo: domain.WechatInfo{
			UnionId: user.WechatUnionId.String,
			OpenId:  user.WechatOpenId.String,
		},
	}
}

func (u *CachedUserRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		Birthday: user.Birthday.UnixMilli(),
		WechatOpenId: sql.NullString{
			String: user.OpenId,
			Valid:  user.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: user.UnionId,
			Valid:  user.UnionId != "",
		},
		AboutMe:  user.AboutMe,
		Nickname: user.Nickname,
		Id:       user.Id,
	}
}

func (u *CachedUserRepository) FindByWechat(ctx context.Context, openid string) (domain.User, error) {
	ue, err := u.dao.FindByWechat(ctx, openid)
	if err != nil {
		return domain.User{}, err
	}
	return u.toDomain(ue), nil
}
