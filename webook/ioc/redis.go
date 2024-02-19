package ioc

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/jayleonc/geektime-go/webook/pkg/redisx"
	"github.com/prometheus/client_golang/prometheus"
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
	opts := prometheus.SummaryOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "redis_req",
		Help:      "统计 redis 响应时间和命中率",
	}
	hook := redisx.NewPrometheusHook(opts)
	redisClint.AddHook(hook)
	return redisClint
}

func InitRLockClient(client redis.Cmdable) *rlock.Client {
	return rlock.NewClient(client)
}
