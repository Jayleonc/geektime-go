package repository

import (
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	mock_dao "github.com/jayleonc/geektime-go/webook/internal/repository/dao/mocks"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCachedArticleRepository_Sync(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "新建同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				adao := mock_dao.NewMockArticleAuthorDAO(ctrl)
				adao.EXPECT().Create(gomock.Any(), dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(int64(1), nil)
				rdao := mock_dao.NewMockArticleReaderDAO(ctrl)
				rdao.EXPECT().
					Upsert(gomock.Any(), dao.Article{
						Id:       1,
						Title:    "我的标题",
						Content:  "我的内容",
						AuthorId: 123,
					}).Return(nil)
				return adao, rdao
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
		{
			name: "修改同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				adao := mock_dao.NewMockArticleAuthorDAO(ctrl)
				adao.EXPECT().Update(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				rdao := mock_dao.NewMockArticleReaderDAO(ctrl)
				rdao.EXPECT().
					Upsert(gomock.Any(), dao.Article{
						Id:       11,
						Title:    "我的标题",
						Content:  "我的内容",
						AuthorId: 123,
					}).Return(nil)
				return adao, rdao
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 11,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//ctrl := gomock.NewController(t)
			//defer ctrl.Finish()
			//authorDAO, readerDAO := tc.mock(ctrl)
			//repo := NewCachedArticleRepository(readerDAO, authorDAO)
			//id, err := repo.Sync(context.Background(), tc.art)
			//assert.Equal(t, tc.wantErr, err)
			//assert.Equal(t, tc.wantId, id)
		})
	}
}
