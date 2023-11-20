package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client   *sms.Client
	appId    string
	signName string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	request := sms.NewSendSmsRequest()
	request.SetContext(ctx) // 链路数据，往下传
	request.SmsSdkAppId = &s.appId
	request.SignName = &s.signName
	request.TemplateId = &tplId
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(numbers)

	response, err := s.client.SendSms(request)
	// 处理异常
	if err != nil {
		return err
	}

	set := response.Response.SendStatusSet
	for i := range set {
		if set[i] == nil {
			// 基本不可能进来这里
			continue
		}
		if *set[i].Code != "Ok" {
			// 循环中，只要有一条短信发送失败，就会直接返回
			// todo 或许应该找出失败的那些，然后重新发送
			return fmt.Errorf("send failed，lua: %v，message: %v\n",
				*set[i].Code, *set[i].Message)
		}
	}
	return nil
}

func (s *Service) toPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data, func(idx int, src string) *string {
		return &src
	})
}

// NewService 依赖注入的形式，只关注如何发送短信，不在乎如何初始化
func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    appId,
		signName: signName,
	}
}
