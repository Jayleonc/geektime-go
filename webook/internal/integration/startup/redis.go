package startup

import (
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	redisClint := redis.NewClient(&redis.Options{
		Addr:     "175.178.58.198:16379",
		Password: "jayleonc",
	})
	return redisClint
}
