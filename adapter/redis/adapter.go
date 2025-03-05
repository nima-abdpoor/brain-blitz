package redis

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type Z struct {
	Score  float64
	Member interface{}
}

type Adapter struct {
	client *redis.Client
}

func New(config Config) *Adapter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})
	log.Info("✅ Redis is up running...")

	return &Adapter{client: rdb}
}

func (a Adapter) Client() *redis.Client {
	return a.client
}

func (a Adapter) ZAdd(ctx context.Context, key string, members ...Z) error {
	var redisZ []redis.Z
	for _, member := range members {
		redisZ = append(redisZ, redis.Z{
			Score:  member.Score,
			Member: member.Member,
		})
	}
	cmd := a.client.ZAdd(ctx, key, redisZ...)
	return cmd.Err()
}

func (a Adapter) ZRange(ctx context.Context, key string, start, stop int64, withScores bool) (error, []Z) {
	var cmd *redis.ZSliceCmd
	if withScores {
		cmd = a.client.ZRangeWithScores(ctx, key, start, stop)
	} else {
		res := a.client.ZRange(ctx, key, start, stop)
		if res.Err() != nil {
			return res.Err(), nil
		} else {
			var zRanges []Z
			for _, value := range res.Val() {
				zRanges = append(zRanges, Z{
					Member: value,
				})
			}
			return nil, zRanges
		}
	}
	if cmd.Err() != nil {
		return nil, nil
	} else {
		var zRanges []Z
		for _, value := range cmd.Val() {
			zRanges = append(zRanges, Z{
				Member: value.Member,
				Score:  value.Score,
			})
		}
		return nil, zRanges
	}
}
