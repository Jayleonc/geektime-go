package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, uid int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	UpdateById(ctx context.Context, entity User) error
	FindByWechat(ctx *gin.Context, openid string) (User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{
		db: db,
	}
}

func (d *GormUserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := d.db.Table(user.TableName()).WithContext(ctx).Create(&user).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (d *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Table(user.TableName()).Where("email = ?", email).Find(&user).Error
	return user, err
}

func (d *GormUserDAO) FindById(ctx context.Context, uid int64) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Table(user.TableName()).Where("id = ?", uid).First(&user).Error
	return user, err
}

func (d *GormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Table(user.TableName()).Where("phone = ?", phone).First(&user).Error
	return user, err
}

func (d *GormUserDAO) UpdateById(ctx context.Context, entity User) error {

	// 这种写法依赖于 GORM 的零值和主键更新特性
	// Update 非零值 WHERE id = ?
	//return dao.db.WithContext(ctx).Updates(&entity).Error
	return d.db.WithContext(ctx).Model(&entity).Where("id = ?", entity.Id).
		Updates(map[string]any{
			"utime":    time.Now().UnixMilli(),
			"nickname": entity.Nickname,
			"birthday": entity.Birthday,
			"about_me": entity.AboutMe,
		}).Error
}

func (d *GormUserDAO) FindByWechat(ctx *gin.Context, openid string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Table(user.TableName()).Where("wechat_open_id = ?", openid).First(&user).Error
	return user, err
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 代表这是一个可以为 NULL 的列
	Email    sql.NullString `gorm:"unique"`
	Password string

	Nickname string `binding:"omitempty" gorm:"type=varchar(128)"`
	// YYYY-MM-DD
	Birthday int64
	AboutMe  string `binding:"omitempty" gorm:"type=varchar(4096)"`

	// 代表这是一个可以为 NULL 的列
	Phone sql.NullString `gorm:"unique"`

	WechatOpenId  sql.NullString
	WechatUnionId sql.NullString

	// 时区，UTC 0 的毫秒数
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}

func (u User) TableName() string {
	return "users"
}
