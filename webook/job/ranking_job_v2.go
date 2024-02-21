package job

import (
	"context"
	"fmt"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type RankingJobV2 struct {
	svc     service.RankingService
	l       logger.Logger
	timeout time.Duration
	client  redis.Cmdable

	lock      *rlock.Lock
	localLock *sync.Mutex
	key       string

	loadKey string // 用于存储节点负载信息的Key
	maxLoad int    // 可接受的最大负载值
}

func NewRankingJobV2(svc service.RankingService, l logger.Logger, timeout time.Duration, client redis.Cmdable) *RankingJobV2 {
	hostname, err := os.Hostname()
	if err != nil {
		return nil
	}
	r := &RankingJobV2{svc: svc, l: l, timeout: timeout, client: client, localLock: &sync.Mutex{}, key: "job:ranking",
		loadKey: "node:load:" + hostname, maxLoad: 50}
	go r.updateLoad()
	return r
}

func (r *RankingJobV2) tryLock(ctx context.Context) (bool, error) {
	// 锁的值可以是一个唯一标识，如UUID，这里使用主机名
	val, err := os.Hostname()
	if err != nil {
		return false, err
	}
	// 尝试获取锁，设置过期时间防止死锁
	success, err := r.client.SetNX(ctx, r.key, val, r.timeout).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (r *RankingJobV2) autoRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 更新锁的过期时间
			r.client.Expire(ctx, r.key, r.timeout)
		}
	}
}

func (r *RankingJobV2) unlock(ctx context.Context) error {
	_, err := r.client.Del(ctx, r.key).Result()
	return err
}

// updateLoad 更新负载信息
func (r *RankingJobV2) updateLoad() {
	for {
		// 生成0到100之间的随机数模拟负载
		load := rand.Intn(101)
		// 无论 loadKey 是否存在或过期，它都会被新的负载值替换，并重新设置过期时间
		_, err := r.client.Set(context.Background(), r.loadKey, load, time.Minute+time.Second*5).Result()
		if err != nil {
			r.l.Warn("更新节点负载失败", logger.Error(err))
		}
		// 每隔一段时间更新一次负载
		time.Sleep(time.Minute)
	}
}

func (r *RankingJobV2) checkLoad() bool {
	loadStr, err := r.client.Get(context.Background(), r.loadKey).Result()
	if err != nil {
		r.l.Warn("获取节点负载信息失败", logger.Error(err))
		return false
	}
	load, err := strconv.Atoi(loadStr)
	if err != nil {
		r.l.Warn("节点负载信息转换失败", logger.Error(err))
		return false
	}
	return load <= r.maxLoad
}

func (r *RankingJobV2) Name() string {
	return "ranking"
}

func (r *RankingJobV2) tryLockWithLoadCheck(ctx context.Context) (bool, error) {
	if !r.checkLoad() {
		return false, fmt.Errorf("当前节点负载过高")
	}
	success, err := r.tryLock(ctx)
	if err != nil || !success {
		return false, err
	}
	// 成功获取锁后，再次检查负载
	if !r.checkLoad() {
		r.unlock(ctx) // 如果负载过高，则释放锁
		return false, fmt.Errorf("获取锁后发现当前节点负载过高")
	}
	return true, nil
}

func (r *RankingJobV2) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	if r.lock == nil {
		success, err := r.tryLockWithLoadCheck(ctx)
		if err != nil || !success {
			r.l.Warn("获取分布式锁失败或节点负载过高", logger.Error(err))
			return err
		}

		defer r.unlock(ctx)
	}

	// 执行任务
	return r.svc.TopN(ctx)
}
