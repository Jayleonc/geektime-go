package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	mock_repository "github.com/jayleonc/geektime-go/webook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("hello#world123")
	password, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	fmt.Println(string(password))
}

func Test_userService_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期输入
		ctx      context.Context
		email    string
		password string

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepository := mock_repository.NewMockUserRepository(ctrl)
				userRepository.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$4IuBgNxUHe6CcgMc1bNTWuiOVBaG5lJUo1NwZlbSnwzm8g5VL7lEa",
						Phone:    "15212345678",
					}, nil)
				return userRepository
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$4IuBgNxUHe6CcgMc1bNTWuiOVBaG5lJUo1NwZlbSnwzm8g5VL7lEa",
				Phone:    "15212345678",
			},
		},
		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepository := mock_repository.NewMockUserRepository(ctrl)
				userRepository.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return userRepository
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepository := mock_repository.NewMockUserRepository(ctrl)
				userRepository.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("db错误"))
				return userRepository
			},
			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  errors.New("db错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepository := mock_repository.NewMockUserRepository(ctrl)
				userRepository.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$4IuBgNxUHe6CcgMc1bNTWuiOVBaG5lJUo1NwZlbSnwzm8g5VL7lEa",
						Phone:    "15212345678",
					}, nil)
				return userRepository
			},
			email:    "123@qq.com",
			password: "hello#world",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			service := NewUserService(repo)
			user, err := service.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
