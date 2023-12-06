package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/integration/startup"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestUserHandler_SendSMSCode(t *testing.T) {
	rdb := startup.InitRedis()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		phone string

		wantCode int
		wantBody ginx.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				key := "phone_code:login:18174715374"
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) > 0)
				duration, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, duration > time.Minute*9+time.Second*5)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
				err = rdb.Del(ctx, key+":cnt").Err()
				assert.NoError(t, err)
			},
			phone:    "18174715374",
			wantCode: http.StatusOK,
			wantBody: ginx.Result{
				Code: 200,
				Msg:  "发送成功",
			},
		},
		{
			name: "发送失败，未输入手机号",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
			},
			phone:    "",
			wantCode: http.StatusOK,
			wantBody: ginx.Result{
				Code: 4,
				Msg:  "请输入手机号",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				key := "phone_code:login:18174715374"
				err := rdb.Set(ctx, key, "123456", time.Minute*9+time.Second*30).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				key := "phone_code:login:18174715374"
				_, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
			},
			phone:    "18174715374",
			wantCode: http.StatusOK,
			wantBody: ginx.Result{
				Code: 400,
				Msg:  "短信发送太频繁，请稍后再试",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			server := startup.InitWebServer()

			request, err := http.NewRequest(http.MethodPost,
				"/users/login_sms/code/send",
				bytes.NewReader([]byte(fmt.Sprintf(`{"phone": "%s"}`, tc.phone))))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, request)

			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res ginx.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)

			assert.Equal(t, tc.wantBody, res)
		})
	}
}
