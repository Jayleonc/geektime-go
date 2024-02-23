package wire

import (
	"github.com/jayleonc/geektime-go/webook/internal/events"
	"github.com/jayleonc/geektime-go/webook/pkg/grpcx"
)

type App struct {
	Consumers []events.Consumer
	Server    *grpcx.Server
}
