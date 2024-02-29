// Package service provides ...
package service

import (
	"context"
	"errors"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或密码不对，请重新输入")
)

type UserService interface {
	Signup(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
	Send(ctx context.Context, biz, phone string) error
}

type userService struct {
	repo repository.UserRepository
}

func (u *userService) Send(ctx context.Context, biz, phone string) error {
	//TODO implement me
	panic("implement me")
}

func (u *userService) Update(ctx context.Context, user domain.User) error {
	return u.repo.UpdateNonZeroFields(ctx, user)
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (u *userService) Signup(ctx context.Context, user domain.User) error {
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)
	return u.repo.Create(ctx, user)
}

func (u *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	return user, nil
}

func (u *userService) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return u.repo.FindById(ctx, uid)
}

func (u *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先查找用户，大部份用户都应是已经存在的用户
	user, err := u.repo.FindByPhone(ctx, phone)
	// 如果 gorm 使用 Find 方法，用户不存在时，err 不会返回 ErrUserNotFound，只会返回 nil
	// 要使用 First 方法，才能正常返回 ErrUserNotFound
	if !errors.Is(err, repository.ErrUserNotFound) {
		return user, err
	}
	zap.L().Info("新用户", zap.String("phone", phone))
	// 如果没有找到用户，注册一个
	err = u.repo.Create(ctx, domain.User{
		Phone: phone,
	})

	// 两者错误情况，一种是唯一索引冲突（phone），另一个中 err!=nil 系统错误
	// 要么 err == nil，要么 ErrDuplicateUser，代表用户存在
	// 注意：正常用户是到不了这里的，到这里应该是遇到并发问题
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return domain.User{}, nil
	}
	// 这里是注册成功之后，再找一次，这里存在主从延迟问题
	// todo 主从延迟，往主表里新增数据后，还来得及同步到从表，就去从表里 find，可能找不到
	// 解决思路：强制走主表查询
	return u.repo.FindByPhone(ctx, phone)

}

func (u *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	user, err := u.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if err != repository.ErrUserNotFound {
		return user, err
	}
	// 这边就是意味着是一个新用户
	// JSON 格式的 wechatInfo
	//svc.logger.Info("新用户", zap.Any("wechatInfo", wechatInfo))
	err = u.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	return u.repo.FindByWechat(ctx, wechatInfo.OpenId)
}
