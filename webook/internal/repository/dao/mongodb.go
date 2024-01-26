package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]Article, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func (m MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) *MongoDBArticleDAO {
	return &MongoDBArticleDAO{
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		node:    node,
	}
}

func (m MongoDBArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Id = m.node.Generate().Int64()
	article.Ctime = now
	article.Utime = now
	_, err := m.col.InsertOne(ctx, article)
	if err != nil {
		return 0, err
	}
	return article.Id, nil
}

func (m MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{"id", art.Id},
		bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		// 创作者不对，说明有人在瞎搞
		return errors.New("ID 不对或者创作者不对")
	}
	return nil
}

func (m MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now
	// liveCol 是 INSERT or Update 语义
	filter := bson.D{bson.E{"id", art.Id},
		bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", art},
		bson.E{"$setOnInsert",
			bson.D{bson.E{"ctime", now}}}}
	_, err = m.liveCol.UpdateOne(ctx,
		filter, set,
		options.Update().SetUpsert(true))
	return id, err
}
