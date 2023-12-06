package ioc

import "github.com/jayleonc/geektime-go/webook/internal/service/oauth/wechat"

func InitWeChatService() wechat.Service {
	appID := "12312312"
	appSecert := "12312312"
	return wechat.NewService(appID, appSecert)
}
