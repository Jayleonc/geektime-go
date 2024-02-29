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
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/auth"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
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
}

func NewSmsService(svc sms.Service, repo repository.AsyncTaskRepository, l logger.Logger) *SmsService {
	return &SmsService{svc: svc, repo: repo, l: l}
}

func (s *SmsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 安全地检查是否跳过异步检查，假设默认不跳过
	skipAsyncCheck := false
	if value, ok := ctx.Value("skipAsyncCheck").(bool); ok {
		skipAsyncCheck = value
	}

	if !skipAsyncCheck && s.needAsync() {
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
	// 记录开始时间
	startTime := time.Now()

	err := s.svc.Send(ctx, tplId, args, numbers...)

	// 计算响应时间并更新 responseTimes
	s.updateResponseTimes(time.Since(startTime))

	// 更新 errors
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

func (s *SmsService) AsyncSend(as async.Sms) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 记录开始时间
	startTime := time.Now()

	// 执行异步发送逻辑
	fmt.Println("异步发送短信")
	ctx = auth.WithSkipAuth(ctx, true)
	ctx = WithSkipAsyncCheck(ctx, true)
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
	// 限制记录数量，例如最近100条
	if len(s.responseTimes) > 100 {
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

// 提前引导你们，开始思考系统容错问题
// 你们面试装逼，赢得竞争优势就靠这一类的东西
func (s *SmsService) needAsync() bool {
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率

	s.mu.Lock()
	defer s.mu.Unlock()

	// 分析响应时间
	if len(s.responseTimes) > 0 {
		avgResponseTime := s.calculateAverageResponseTime()
		if avgResponseTime > 500*time.Millisecond { // 使用绝对阈值
			return true
		}
	}

	// 分析错误率
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

// WithSkipAsyncCheck 创建一个新的context，包含一个标记以跳过异步检查。
func WithSkipAsyncCheck(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, "skipAsyncCheck", skip)
}
