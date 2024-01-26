package startup

import (
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	redisClint := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "jayleonc",
	})
	return redisClint
}
