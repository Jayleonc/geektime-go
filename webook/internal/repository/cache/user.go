package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
	Del(ctx context.Context, id int64) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable // 面向接口编程，有多种实现，比如 ClusterClient（集群），Client 等
	expiration time.Duration
}

// NewUserCache 松耦合的设计，不要自己初始化 redis，让外面传进来，就能实现通用啦
func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := c.key(uid)
	// 使用 JSON 序列化
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal([]byte(data), &user)
	return user, err
}

func (c *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := c.key(du.Id)
	// 使用 JSON 序列化
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisUserCache) Del(ctx context.Context, id int64) error {
	return c.cmd.Del(ctx, c.key(id)).Err()
}

func (c *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}
