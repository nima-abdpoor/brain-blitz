// Package matchqueue shares queue presence between game sockets and matchmaking.
package matchqueue

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var categories = []string{"SPORT", "MUSIC", "TECH"}

const presencePrefix = "queue_presence:"

func keys(prefix, id string) []string {
	k := []string{presencePrefix + id}
	for _, c := range categories {
		k = append(k, prefix+":"+c)
	}
	return k
}

func Join(ctx context.Context, client *redis.Client, id, category string) error {
	return client.Set(ctx, presencePrefix+id, category, 20*time.Minute).Err()
}

func Active(ctx context.Context, client *redis.Client, id, category string) bool {
	v, err := client.Get(ctx, presencePrefix+id).Result()
	return err == nil && v == category
}

func Leave(ctx context.Context, client *redis.Client, prefix, id string) error {
	return client.Eval(ctx, `redis.call('DEL',KEYS[1]); for i=2,#KEYS do redis.call('ZREM',KEYS[i],ARGV[1]) end; return 1`, keys(prefix, id), id).Err()
}

// A delayed Kafka join cannot reinsert a cancelled player. Rejoining a
// different category removes the previous category entry atomically.
func Enqueue(ctx context.Context, client *redis.Client, prefix, id, category string) error {
	return client.Eval(ctx, `if redis.call('GET',KEYS[1]) ~= ARGV[2] then return 0 end; for i=2,#KEYS do redis.call('ZREM',KEYS[i],ARGV[1]) end; redis.call('ZADD',ARGV[3],ARGV[4],ARGV[1]); return 1`, keys(prefix, id), id, category, prefix+":"+category, time.Now().UnixMicro()).Err()
}
