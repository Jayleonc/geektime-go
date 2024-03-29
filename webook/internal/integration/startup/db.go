package startup

import (
	dao2 "github.com/jayleonc/geektime-go/webook/interactive/repository/dao"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:jayleonc@tcp(localhost:3306)/geektime"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
