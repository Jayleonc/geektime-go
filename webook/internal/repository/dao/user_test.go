package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	mysql_driver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGormUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string
		mock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				res := sqlmock.NewResult(123, 1)
				//  sql 正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnResult(res)
				return db
			},
			ctx:  context.Background(),
			user: User{},
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				//  sql 正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(&mysql_driver.MySQLError{Number: 1062})
				return db
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: ErrDuplicateEmail,
		},
		{
			name: "插入失败",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				//  sql 正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(errors.New("发生错误"))
				return db
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: errors.New("发生错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.mock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			assert.NoError(t, err)
			dao := NewUserDAO(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
