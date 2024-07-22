package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func NewMongoDB(config Config) (*mongo.Collection, error) {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("%s://%s:%v", config.User, config.Host, config.Port))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Println(fmt.Sprintf("connecting to mongoDB failed %v\n", err))
		return nil, err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Println(fmt.Sprintf("ping mongoDB connection failed %v\n", err))
		return nil, err
	}
	collection := client.Database("BrainBlitz").Collection("access_control")
	if err != nil {
		log.Println(fmt.Sprintf("failed to get collection %v\n", err))
		return nil, err
	}

	createAccessControlData(collection)
	return collection, nil
}

func createAccessControlData(db *mongo.Collection) {
	result, err := db.InsertOne(context.Background(), &AccessControl{
		RoleType:    "admin",
		Permissions: []string{"USER_LIST", "USER_DELETE"},
	})
	fmt.Println(result)
	if err != nil {
		log.Println(fmt.Sprintf("failed to insert Data %v\n", err))
	}
}

type Config struct {
	User string `koanf:"user"`
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}
