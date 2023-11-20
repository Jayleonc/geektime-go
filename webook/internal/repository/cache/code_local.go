package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache  *sync.Map
	expire time.Duration // 验证码有效时间，固定 10 分钟
}

// codeEntry 表示存储在缓存中的验证码信息
type codeEntry struct {
	code      string
	timestamp time.Time // 存储验证码时间信息
	cnt       int32     // 可验证次数
}

func (lc *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	// 该实现存储 key 的实际存储时的时间，验证时使用当前时间减去存储时间，对比是否超过验证码有效时间 10 分钟
	fmt.Println("使用本地缓存发送验证码咯～～")
	key := lc.generateKey(biz, phone)
	val, ok := lc.cache.Load(key)
	if !ok { // 大部份直接进这里，都是第一次发，还没有验证码
		lc.cache.Store(key, codeEntry{code: code, timestamp: time.Now(), cnt: 3})
		fmt.Println("第一次发送验证码，成功啦～")
		return nil
	}

	// 如果拿到了验证码，代表当前是重发验证码
	entry, ok := val.(codeEntry)
	if !ok {
		return errors.New("系统错误")
	}

	t := time.Since(entry.timestamp)
	fmt.Println("当前度过时间：", t)
	if t <= time.Minute {
		// 不到一分钟，发送第二次验证码
		return ErrCodeSendTooMany
	}

	// 重发
	lc.cache.Store(key, codeEntry{
		code:      code,
		timestamp: time.Now(),
		cnt:       3,
	})

	return nil
}

func (lc *LocalCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	fmt.Println("使用本地缓存验证验证码咯～～")
	key := lc.generateKey(biz, phone)
	val, loaded := lc.cache.Load(key)
	if !loaded {
		// 没发验证码
		return false, ErrKeyNotExist
	}

	entry, ok := val.(codeEntry)
	if !ok {
		return false, errors.New("系统错误")
	}

	if entry.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	entry.cnt--
	lc.cache.Store(key, entry)

	if entry.code != code {
		return false, nil
	}

	// 验证通过后，检查验证码是否过期
	if time.Since(entry.timestamp) > lc.expire {
		return false, ErrCodeExpired
	}

	return true, nil
}

func (lc *LocalCodeCache) generateKey(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

// NewLocalCodeCache 创建一个新的 LocalCodeCache 实例
func NewLocalCodeCache() CodeCache {
	return &LocalCodeCache{
		cache:  new(sync.Map),
		expire: time.Minute * 10,
	}
}
