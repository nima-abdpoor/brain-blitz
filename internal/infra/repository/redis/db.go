package redis

import (
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

func ZRange(ctx context.Context, client *redis.Client, key, min, max string) ([]redis.Z, error) {
	const op = "Redis.ZRange"
	if result, err := client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result(); err != nil {
		logger.Logger.Named(op).Error("error in ZRangeByScore", zap.Error(err))
		return nil, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	} else {
		return result, nil
	}
}
