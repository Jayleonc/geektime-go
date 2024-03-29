package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/events"
	"github.com/jayleonc/geektime-go/webook/pkg/retry"
	"github.com/robfig/cron/v3"
)

type App struct {
	Web       *gin.Engine
	Consumers []events.Consumer
	Corn      *cron.Cron
	Scheduler *retry.Scheduler
}
