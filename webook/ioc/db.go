package ioc

import (
	"github.com/jayleonc/geektime-go/webook/internal/config"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&dao.User{})
	return db
}
