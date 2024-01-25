package startup

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func InitMongoDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}
	opts := options.Client().
		ApplyURI("mongodb://root:Jayleonc@175.178.58.198:27017/").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}
	return client.Database("webook")
}
