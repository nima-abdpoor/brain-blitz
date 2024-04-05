package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
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

func main() {
	adapter := New(Config{
		Host:     "localhost",
		Port:     6380,
		Password: "",
		DB:       0,
	})

	status, err := adapter.Client().Set(context.Background(), "key1", "two", time.Second*120).Result()
	if err != nil {
		fmt.Println(err)
	}

	//status, err := adapter.Client().Get(context.Background(), "key1").Result()
	//if err != nil {
	//	fmt.Println(err)
	//}

	_, err = adapter.client.ZAdd(
		context.Background(),
		fmt.Sprintf("%s:%s", "waitingList", "footbal"),
		redis.Z{Score: float64(time.Now().UnixMicro()), Member: 1}).Result()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(status)
}
