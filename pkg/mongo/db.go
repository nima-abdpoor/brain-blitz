package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	DB *mongo.Database
}

func NewDB(config Config, ctx context.Context) (*Database, error) {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("%s://%s:%v", config.User, config.Host, config.Port))
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return nil, err
	}

	database := client.Database(config.Name)
	return &Database{
		DB: database,
	}, nil
}

func Ping(db *mongo.Database, ctx context.Context) error {
	return db.Client().Ping(ctx, nil)
}

func Close(db *mongo.Database, ctx context.Context) error {
	return db.Client().Disconnect(ctx)
}
