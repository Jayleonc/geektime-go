package ioc

import (
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var c Config
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		Logger: gormlogger.New(log.New(os.Stdout, "\n", log.LstdFlags), gormlogger.Config{
			SlowThreshold: time.Second,
			Colorful:      true,
			LogLevel:      gormlogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&dao.User{})
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Val: i})
}
