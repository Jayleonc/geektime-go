package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/integration/startup"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (s *ArticleHandlerSuite) SetupSuite() {
	s.db = startup.InitDB()
	hdl := startup.InitArticleHandler(dao.NewArticleGORMDAO(s.db))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{
			Uid: 123,
		})
	})
	hdl.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleHandlerSuite) TearDownTest() {
	s.db.Exec("truncate table `articles`")
}

func (s *ArticleHandlerSuite) TestArticle_Publish() {
	t := s.T()

	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				s.db.Where("author_id = ?", 123).First(&art)
				assert.Equal(t, "hello，你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, uint8(2), art.Status)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt dao.PublishedArticle
				s.db.Where("author_id = ?", 123).First(&publishedArt)
				assert.Equal(t, "hello，你好", publishedArt.Title)
				assert.Equal(t, "随便试试", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
				assert.Equal(t, uint8(2), publishedArt.Status)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
				Code: http.StatusOK,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				s.db.Create(&dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   1,
					Utime:    234,
					AuthorId: 123,
				})
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				s.db.Where("id = ?", 2).First(&art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, uint8(2), art.Status)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)
				var publishedArt dao.PublishedArticle
				s.db.Where("id = ?", 2).First(&publishedArt)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.Equal(t, uint8(2), publishedArt.Status)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
				Code: http.StatusOK,
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   1,
					Utime:    234,
					AuthorId: 123,
				}
				s.db.Create(&art)
				part := dao.PublishedArticle(art)
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				var art dao.Article
				s.db.Where("id = ?", 3).First(&art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, uint8(2), art.Status)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)

				var part dao.PublishedArticle
				s.db.Where("id = ?", 3).First(&part)
				assert.Equal(t, "新的标题", part.Title)
				assert.Equal(t, "新的内容", part.Content)
				assert.Equal(t, int64(123), part.AuthorId)
				assert.Equal(t, uint8(2), part.Status)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.True(t, part.Utime > 234)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 3,
				Code: http.StatusOK,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					Status:  1,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
				}
				s.db.Create(&art)
				part := dao.PublishedArticle(dao.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   2,
					Utime:    234,
					AuthorId: 789,
				})
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				var art dao.Article
				s.db.Where("id = ?", 4).First(&art)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, uint8(1), art.Status)
				assert.Equal(t, int64(789), art.AuthorId)

				var part dao.PublishedArticle
				// 数据没有变化
				s.db.Where("id = ?", 4).First(&part)
				assert.Equal(t, "我的标题", part.Title)
				assert.Equal(t, "我的内容", part.Content)
				assert.Equal(t, int64(789), part.AuthorId)
				assert.Equal(t, uint8(2), part.Status)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.Equal(t, int64(234), part.Utime)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

func (s *ArticleHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		art Article

		wantCode int
		wantRes  Result[int64]
	}{

		{
			name:   "新建文章",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				// 验证确实写到数据库里了
				var art dao.Article
				err := s.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "文章内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
			},
			art: Article{
				Title:   "我的标题",
				Content: "文章内容",
			},

			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: http.StatusOK,
				Data: 1,
			},
		},
		{
			name: "修改文章",
			before: func(t *testing.T) {
				article := dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   1,
					Ctime:    456,
					Utime:    789,
				}
				err := s.db.Create(article).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证确实写到数据库里了
				var art dao.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 789)
				assert.True(t, art.Id == 2)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, uint8(1), art.Status)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},

			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: http.StatusOK,
				Data: 2,
			},
		},
		{
			name: "修改文章 - 修改了别人的文章",
			before: func(t *testing.T) {
				// 创建一篇文章
				article := dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   1,
					Ctime:    456,
					Utime:    789,
				}
				err := s.db.Create(article).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据没有变
				var art dao.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id == 3)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(234), art.AuthorId)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, uint8(1), art.Status)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},

			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg: "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewReader(reqBody))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)

			assert.Equal(t, tc.wantCode, recorder.Code)

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestArticleHandler(t *testing.T) {
	suite.Run(t, &ArticleHandlerSuite{})
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64
	Title   string
	Content string
}
