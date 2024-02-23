package startup

import "github.com/jayleonc/geektime-go/webook/pkg/logger"

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
