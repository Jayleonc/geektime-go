package producer

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/pkg/migrator/events"
)

type Producer interface {
	ProducerInconsistentEvent(ctx context.Context, evt events.InconsistentEvent) error
}
