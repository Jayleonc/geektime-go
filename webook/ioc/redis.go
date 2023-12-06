package ioc

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		Addr     string
		Password string
	}
	c := Config{}
	err := viper.UnmarshalKey("redis", &c)
	if err != nil {
		panic(err)
	}
	redisClint := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
	})
	err = redisClint.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}
	return redisClint
}
