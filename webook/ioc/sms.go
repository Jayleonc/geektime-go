package ioc

import (
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/async"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/localsms"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/prometheus"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/tencent"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
)

func InitSMSService() sms.Service {
	//ratelimit.NewRateLimitSMSService(localsms.NewService(), limiter.NewRedisSlidingWindowLimiter())
	service := localsms.NewService()
	opts := prometheus2.SummaryOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "sms_req",
		Help:      "统计 sms 请求响应时间",
	}
	decorator := prometheus.NewSMSDecorator(service, opts)
	return decorator
}

func InitAsyncSMSService(repo repository.AsyncTaskRepository, l logger.Logger) *async.SmsService {
	// 首先，初始化装饰过的SMS服务
	decoratedService := InitSMSService()

	// 然后，使用装饰过的服务初始化asyncSmsService
	asyncSmsService := async.NewSmsService(decoratedService, repo, l)

	return asyncSmsService
}

func initTencentSMSService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯 SMS 的 secret key")
	}
	c, err := tencentSMS.NewClient(
		common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile(),
	)
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, "1400842696", "妙影科技")
}
