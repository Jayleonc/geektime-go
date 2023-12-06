package cache

import (
	"context"
	"errors"
	"fmt"
	mockv9 "github.com/jayleonc/geektime-go/webook/internal/repository/cache/redismocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) redis.Cmdable
		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := mockv9.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(1))
				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("login", "18174715374")}, []any{"123456"}).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "18174715374",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "返回error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := mockv9.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis错误"))
				cmd.SetVal(int64(1))
				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("login", "18174715374")}, []any{"123456"}).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "18174715374",
			code:    "123456",
			wantErr: errors.New("redis错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := NewCodeCache(tc.mock(ctrl))
			err := cache.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}

}