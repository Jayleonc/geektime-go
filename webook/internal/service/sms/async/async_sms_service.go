package async

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/domain/async"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type SmsService struct {
	svc  sms.Service
	repo repository.AsyncTaskRepository
	l    logger.Logger

	responseTimes []time.Duration // 响应时间记录
	errors        []bool          // 错误记录
	mu            sync.Mutex      // 保护 responseTimes 和 errors

	asyncMode      bool
	asyncStartTime time.Time
}

func NewSmsService(svc sms.Service, repo repository.AsyncTaskRepository, l logger.Logger) *SmsService {
	s := &SmsService{svc: svc, repo: repo, l: l}
	//startPerformanceMonitoring(s)
	return s
}

func (s *SmsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 检查是否跳过异步检查，如果设置了跳过异步，则直接同步发送，只在转异步的 AsyncSend 设置了，避免洋葱导致的循环异步
	if skipAsyncCheck, _ := ctx.Value(sms.SkipAsyncCheck).(bool); skipAsyncCheck {
		return s.directSend(ctx, tplId, args, numbers...)
	}

	// 检查是否触发了限流
	asyncModeCheck := false
	if value, ok := ctx.Value(sms.AsyncMode).(bool); ok {
		// 触发限流，asyncModeCheck = true
		asyncModeCheck = value
	}

	//var mu sync.RWMutex
	//mu.RLock() // 在读取 asyncMode 前加锁
	//asyncModeActive := s.asyncMode
	//mu.RUnlock() // 读取完毕后立即解锁

	// 使用 needAsync，即时性强，能快速切换到异步，减少延迟和失误率
	needAsync := s.needAsync() || asyncModeCheck
	// 使用 asyncMode，避免每次执行发送操作都执行 needAsync，减少 CPU 的使用和性能开销
	//needAsync := asyncModeActive || asyncModeCheck

	if needAsync {
		Sms := async.Sms{
			TplId:   tplId,
			Args:    args,
			Numbers: numbers,
		}

		parameters, err := json.Marshal(Sms)
		if err != nil {
			return err
		}
		task := service.NewRetryTask("SMS", "SMS", string(parameters), 5)
		// 存储任务到数据库，调度器会异步执行改任务。
		storeErr := s.repo.StoreTask(ctx, task)
		return storeErr
	}

	return s.directSend(ctx, tplId, args, numbers...)
}

// directSend 封装了直接发送短信的逻辑，以避免重复代码
func (s *SmsService) directSend(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 记录开始时间
	startTime := time.Now()

	// 执行发送操作
	err := s.svc.Send(ctx, tplId, args, numbers...)

	// 计算响应时间并更新
	s.updateResponseTimes(time.Since(startTime))

	// 更新错误状态
	s.updateErrors(err != nil)

	return err
}

func (s *SmsService) Execute() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 如果能读取到 Sms 的 task
	tasks, err := s.repo.LoadTasks(ctx, "SMS")
	if err != nil {
		return err
	}

	// 没有任务需要处理
	if len(tasks) == 0 {
		return nil
	}

	// 理论上，不会有很多短信任务，所以不用开协程
	// 但是如果是运营商批量发送那种，就有可能有很多很多短信任务
	for _, task := range tasks {
		var Sms async.Sms
		if err := json.Unmarshal([]byte(task.Parameters), &Sms); err != nil {
			s.l.Error("反序列化失败", logger.Error(err), logger.String("id", task.Id))
			continue // 出错则处理下一个任务
		}

		// 处理每个任务，尝试发送SMS
		s.handleTask(ctx, task, Sms)
	}

	return nil
}

func (s *SmsService) handleTask(ctx context.Context, task async.Task, sms async.Sms) {
	var err error
	isSuccess := false // 默认为失败

	// 在这里不修改task.RetryCount
	// 记录尝试的次数
	attemptsMade := 0
	for attemptsMade < task.RetryCount {
		err = s.AsyncSend(sms)
		if err == nil {
			isSuccess = true // 成功发送
			break
		}
		// 重试之前等待一段时间，重试间隔
		time.Sleep(s.calculateBackoff(attemptsMade))
		attemptsMade++
	}

	if err != nil {
		task.ErrorMessage = err.Error()
	}

	// 更新任务状态
	s.updateTaskStatus(ctx, task, isSuccess, attemptsMade)
}

// 注意，我们添加了一个新参数来正确反映尝试次数
func (s *SmsService) updateTaskStatus(ctx context.Context, task async.Task, isSuccess bool, attemptsMade int) {
	if isSuccess {
		task.Status = int(dao.StatusSuccess)
		// 可以记录实际的尝试次数，而不是修改RetryCount
	} else {
		task.Status = int(dao.StatusFailed)
		// 处理失败逻辑，但不修改RetryCount
	}

	// 假设repo.UpdateTask能够处理状态更新
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		s.l.Error("更新任务状态失败", logger.Error(err), logger.String("id", task.Id), logger.String("status", string(rune(task.Status))))
	}
}

// calculateBackoff 根据重试次数计算重试间隔，这里使用简单的线性退避
func (s *SmsService) calculateBackoff(attempt int) time.Duration {
	return time.Second * time.Duration(attempt+1)
}

func (s *SmsService) checkAndSwitchModes() {
	var mu sync.RWMutex
	mu.Lock()
	defer mu.Unlock()
	// 检查是否应该进入异步模式
	if !s.asyncMode && s.needAsync() {
		s.asyncMode = true
		s.asyncStartTime = time.Now()
		fmt.Println("定时任务，转为异步发送")
		return
	}

	// 检查是否应该退出异步模式
	if s.asyncMode {
		// 检查是否已经过了N分钟
		if time.Since(s.asyncStartTime) > time.Minute*5 {
			// 逐步增加同步处理的请求比例
			// 可以通过控制一个随机数来决定每个请求是同步还是异步处理
			if rand.Float64() < 0.01 { // 保留1%的流量进行同步发送
				s.asyncMode = false
				fmt.Println("定时任务，转为同步发送")
			}
			// 继续监控响应时间和错误率
			if !s.needAsync() {
				s.asyncMode = false
				fmt.Println("定时任务，转为同步发送")
			}
		}
	}
}

func startPerformanceMonitoring(s *SmsService) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.checkAndSwitchModes()
		}
	}()
}

func (s *SmsService) AsyncSend(as async.Sms) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 记录开始时间
	startTime := time.Now()

	// 执行异步发送逻辑
	fmt.Println("异步发送短信")
	ctx = sms.WithSkipAuth(ctx, true)
	ctx = sms.WithSkipAsyncCheck(ctx, true)
	err := s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)

	// 计算响应时间并更新 responseTimes
	s.updateResponseTimes(time.Since(startTime))

	// 更新 errors
	s.updateErrors(err != nil)

	return err
}

func (s *SmsService) updateResponseTimes(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responseTimes = append(s.responseTimes, duration)
	// 限制记录数量，例如最近1000条
	if len(s.responseTimes) > 1000 {
		s.responseTimes = s.responseTimes[1:]
	}
}

func (s *SmsService) updateErrors(errOccurred bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors = append(s.errors, errOccurred)
	// 同样限制记录数量
	if len(s.errors) > 100 {
		s.errors = s.errors[1:]
	}
}

// needAsync
func (s *SmsService) needAsync() bool {
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1.检测到响应时间的突然增加，直接转为异步
	if s.checkResponseTimeIncrease() {
		return true
	}

	// 2.使用动态阈值，分析响应时间
	threshold := s.calculateResponseTimeThreshold()
	if len(s.responseTimes) > 0 {
		avgResponseTime := s.calculateAverageResponseTime()
		if avgResponseTime > threshold {
			return true
		}
	}

	// 3.分析错误率
	if len(s.errors) > 0 {
		errorRate := s.calculateErrorRate()
		if errorRate > 0.1 { // 假设10%的错误率是不可接受的
			return true
		}
	}
	return false
}

// calculateAverageResponseTime 计算平均响应时间
func (s *SmsService) calculateAverageResponseTime() time.Duration {
	var sum time.Duration
	for _, rt := range s.responseTimes {
		sum += rt
	}
	return sum / time.Duration(len(s.responseTimes))
}

// calculateErrorRate 计算错误率
func (s *SmsService) calculateErrorRate() float64 {
	errorCount := 0
	for _, e := range s.errors {
		if e {
			errorCount++
		}
	}
	return float64(errorCount) / float64(len(s.errors))
}

// 计算动态阈值
func (s *SmsService) calculateResponseTimeThreshold() time.Duration {
	// 计算过去所有响应时间的95%分位数作为阈值
	// 实际实现时需要根据响应时间分布来计算
	return percentile95(s.responseTimes)
}

// 95%分位数意味着95%的数据点都位于该值以下。计算分位数通常需要对数据进行排序
func percentile95(responseTimes []time.Duration) time.Duration {
	length := len(responseTimes)
	if length == 0 {
		return 0
	}
	// 对响应时间进行排序
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})
	// 计算95%分位数的索引
	index := int(0.95 * float64(length-1))
	return responseTimes[index]
}

// 检测响应时间的突然增加
func (s *SmsService) checkResponseTimeIncrease() bool {
	// 比较最后三次响应时间的增长率
	// 实际实现时需要根据具体需求调整逻辑
	return checkIncreaseRate(s.responseTimes)
}

// 检查最后三次请求的平均响应时间是否比前三次有明显增加
func checkIncreaseRate(responseTimes []time.Duration) bool {
	length := len(responseTimes)
	if length < 6 { // 需要至少6个数据点来比较
		return false
	}
	// 分别计算最近三次和前三次响应时间的平均值
	var sumRecent, sumPrevious time.Duration
	for i := length - 3; i < length; i++ {
		sumRecent += responseTimes[i]
	}
	for i := length - 6; i < length-3; i++ {
		sumPrevious += responseTimes[i]
	}
	avgRecent := sumRecent / 3
	avgPrevious := sumPrevious / 3

	// 检查最近三次平均响应时间是否有明显增加
	// 这里检查是否至少增加了40%
	increaseRate := float64(avgRecent-avgPrevious) / float64(avgPrevious)
	return increaseRate > 0.4
}
