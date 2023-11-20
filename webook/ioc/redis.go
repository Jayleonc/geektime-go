package ioc

import (
	"github.com/jayleonc/geektime-go/webook/internal/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	redisClint := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
	})
	return redisClint
}
