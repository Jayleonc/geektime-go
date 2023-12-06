package web

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	mock_service "github.com/jayleonc/geektime-go/webook/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBuilder func(t *testing.T) *http.Request

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mock_service.NewMockUserService(ctrl)
				// EXPECT() 用于设置对模拟对象方法调用的期望的方法
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil)
				codeSvc := mock_service.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world123",
"confirmPassword": "hello#world123"	
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mock_service.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("返回任意一个错误"))
				codeSvc := mock_service.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world123",
"confirmPassword": "hello#world123"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: "系统错误", // 在 mock 匿名函数中，期待 return “返回任意一个错误”，最终响应的 body 是 “系统错误”
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mock_service.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrDuplicateEmail)
				codeSvc := mock_service.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email":"123@qq.com",
"password": "hello#world123",
"confirmPassword": "hello#world123"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突，请换一个",
		},
		{
			name: "Bind错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mock_service.NewMockUserService(ctrl)
				codeSvc := mock_service.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email":"123@qq.com",
"password": "hello#world123"
"confirmPassword": "hello#world123"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: "系统错误",
		},

		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mock_service.NewMockUserService(ctrl)
				codeSvc := mock_service.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email":"123@",
"password": "hello#world123",
"confirmPassword": "hello#world123"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "邮箱格式错误",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mock_service.NewMockUserService(ctrl)
				codeSvc := mock_service.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world",
"confirmPassword": "hello#world"	
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "密码格式错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService, codeService := tc.mock(ctrl)
			hdl := NewUserHandler(userService, codeService)

			server := gin.Default()
			hdl.RegisterRoutes(server)

			builder := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, builder)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestUserEmailPattern(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool // 预期是否匹配
	}{
		{name: "不带@", email: "123456qq.com", match: false},
		{name: "带@ 但没有后缀", email: "123456@qq", match: false},
		{name: "合法email", email: "Jayleonc@163.com", match: true},
	}

	handler := NewUserHandler(nil, nil)

	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			match, err := handler.emailRegexp.MatchString(v.email)
			require.NoError(t, err)
			assert.Equal(t, v.match, match)
		})
	}
}
