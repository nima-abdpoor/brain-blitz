package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type Adapter struct {
	client *redis.Client
}

func New(config Config) *Adapter {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &Adapter{
		client: redisDB,
	}
}

func (a Adapter) Client() *redis.Client {
	return a.client
}

func ZAdd(client *redis.Client, key string, score float64, member interface{}) error {
	_, err := client.ZAdd(
		context.Background(),
		key,
		redis.Z{Score: score, Member: member}).Result()
	return err
}
