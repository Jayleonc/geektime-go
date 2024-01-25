package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	mock_service "github.com/jayleonc/geektime-go/webook/internal/service/mocks"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody  string
		wantCode int
		wantRes  ginx.Response
	}{
		{
			name: "新建文章，发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mock_service.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)

				return svc
			},
			reqBody: `
{
	"title": "我的标题",
	"content": "我的内容"
}
`,
			wantCode: 200,
			wantRes: ginx.Response{
				Code: 200,
				Data: float64(1),
			},
		},
		{
			name: "发表失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mock_service.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("发表失败"))

				return svc
			},
			reqBody: `
{
	"title": "我的标题",
	"content": "我的内容"
}
`,
			wantCode: 200,
			wantRes: ginx.Response{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "Bind failed",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mock_service.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
{
	"title": "我的标题",
	"content": "我的内容"adf
}
`,
			wantCode: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			handler := NewArticleHandler(logger.NewNopLogger(), svc)

			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", ijwt.UserClaims{
					Uid: 123,
				})
			})

			handler.RegisterRoutes(server)

			request, err := http.NewRequest(http.MethodPost,
				"/articles/publish",
				bytes.NewBufferString(tc.reqBody))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, request)

			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var res ginx.Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
