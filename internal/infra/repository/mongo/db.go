package mongo

import (
	"BrainBlitz.com/game/logger"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func NewMongoDB(config Config) (*mongo.Database, error) {
	const op = "mongo.NewMongoDB"
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("%s://%s:%v", config.User, config.Host, config.Port))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logger.Logger.Named(op).Error("connecting to mongoDB failed", zap.Error(err))
		return nil, err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		logger.Logger.Named(op).Error("ping mongoDB connection failed", zap.Error(err))
		return nil, err
	}
	//todo add this to config
	database := client.Database("BrainBlitz")
	createAccessControlData(database)
	return database, nil
}

func createAccessControlData(db *mongo.Database) {
	const op = "mongo.createAccessControlData"
	result, err := db.Collection("access_control").InsertOne(context.Background(), &AccessControl{
		RoleType:    "admin",
		Permissions: []string{"USER_LIST", "USER_DELETE"},
	})
	fmt.Println(result)
	if err != nil {
		logger.Logger.Named(op).Error("failed to seed Data", zap.Error(err))
	}
}

type Config struct {
	User string `koanf:"user"`
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}
