package startup

import "github.com/jayleonc/geektime-go/webook/internal/service/oauth/wechat"

func InitWeChatService() wechat.Service {
	return wechat.NewService("", "")
}
