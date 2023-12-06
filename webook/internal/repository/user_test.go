package repository

import (
	"context"
	"database/sql"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/cache"
	mock_cache "github.com/jayleonc/geektime-go/webook/internal/repository/cache/mocks"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	mock_dao "github.com/jayleonc/geektime-go/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)

		ctx context.Context
		uid int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := mock_cache.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := mock_dao.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: 1199323,
					AboutMe:  "test",
					Phone: sql.NullString{
						String: "18174715374",
						Valid:  true,
					},
					Ctime: 101,
					Utime: 102,
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: time.UnixMilli(1199323),
					AboutMe:  "test",
					Phone:    "18174715374",
					Ctime:    time.UnixMilli(101),
					Utime:    time.UnixMilli(102),
				}).Return(nil)
				return d, c
			},
			uid: 123,
			ctx: context.Background(),

			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(1199323),
				AboutMe:  "test",
				Phone:    "18174715374",
				Ctime:    time.UnixMilli(101),
				Utime:    time.UnixMilli(102),
			},
			wantErr: nil,
		},

		{
			name: "查找成功，缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := mock_cache.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: time.UnixMilli(1199323),
					AboutMe:  "test",
					Phone:    "18174715374",
					Ctime:    time.UnixMilli(101),
					Utime:    time.UnixMilli(102),
				}, nil)
				d := mock_dao.NewMockUserDAO(ctrl)
				return d, c
			},
			uid: 123,
			ctx: context.Background(),

			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(1199323),
				AboutMe:  "test",
				Phone:    "18174715374",
				Ctime:    time.UnixMilli(101),
				Utime:    time.UnixMilli(102),
			},
			wantErr: nil,
		},

		{
			name: "发生错误，未找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := mock_cache.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)
				d := mock_dao.NewMockUserDAO(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{}, dao.ErrRecordNotFound)
				return d, c
			},
			uid: 123,
			ctx: context.Background(),

			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userDAO, userCache := tc.mock(ctrl)

			repository := NewCachedUserRepository(userDAO, userCache)
			user, err := repository.FindById(tc.ctx, tc.uid)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
