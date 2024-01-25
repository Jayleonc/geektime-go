package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestMongo(t *testing.T) {

	ctx := context.TODO()
	// 创建一个监视器
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			fmt.Printf("Command started: %v\n", cse.Command)
		},
	}

	clientOptions := options.Client().
		ApplyURI("mongodb://root:Jayleonc@175.178.58.198:27017")
	clientOptions.SetMonitor(monitor)
	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	assert.NoError(t, err)

	collection := client.Database("webook").Collection("articles")

	one, err := collection.InsertOne(ctx, Article{
		Id:      1,
		Title:   "Title",
		Content: "Content",
	})
	assert.NoError(t, err)
	oid := one.InsertedID.(primitive.ObjectID)
	fmt.Printf("插入ID:%v\n", oid)

	filter := bson.M{
		"id": 1,
	}
	findOne := collection.FindOne(ctx, filter)

	assert.NoError(t, findOne.Err())
	var art Article
	err = findOne.Decode(&art)
	assert.NoError(t, err)

	updateFilter := bson.D{bson.E{"id", 1}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title": "新的标题",
	}}}
	updateOneRes, err := collection.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	updateManyRes, err := collection.UpdateMany(ctx, updateFilter,
		bson.D{bson.E{Key: "$set",
			Value: Article{Content: "新的内容"}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	//deleteFilter := bson.D{bson.E{Key: "id", Value: 1}}
	//delRes, err := collection.DeleteMany(ctx, deleteFilter)
	//assert.NoError(t, err)
	//t.Log("删除文档数量", delRes.DeletedCount)
}

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
