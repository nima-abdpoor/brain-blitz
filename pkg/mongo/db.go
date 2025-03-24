package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type Database struct {
	DB *mongo.Database
}

func NewDB(config Config, ctx context.Context) (*Database, error) {
	if len(config.Hosts) == 0 || len(config.Ports) == 0 || len(config.Hosts) != len(config.Ports) {
		return nil, fmt.Errorf("invalid MongoDB configuration: mismatched hosts and ports")
	}

	var hosts []string
	for i, host := range config.Hosts {
		hosts = append(hosts, fmt.Sprintf("%s:%d", host, config.Ports[i]))
	}

	uri := fmt.Sprintf("mongodb://%s/%s", strings.Join(hosts, ","), config.Name)
	if len(hosts) > 1 {
		uri += fmt.Sprintf("?replicaSet=%s", config.ReplicationName)
	}

	clientOptions := options.Client().ApplyURI(uri)
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
