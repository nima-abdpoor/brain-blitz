// Package matchqueue shares queue presence between game sockets and matchmaking.
package matchqueue

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

type Choice struct {
	Seconds  int
	OptionID string
}

func presence(category string, choices []Choice) string {
	if len(choices) == 0 || choices[0].Seconds == 0 {
		return category
	}
	return fmt.Sprintf("%s|%d|%s", category, choices[0].Seconds, choices[0].OptionID)
}
func ChoiceFor(ctx context.Context, client *redis.Client, id string) (Choice, error) {
	value, err := client.Get(ctx, presencePrefix+id).Result()
	if err != nil {
		return Choice{}, err
	}
	fields := strings.Split(value, "|")
	if len(fields) == 1 {
		return Choice{}, nil
	}
	if len(fields) != 3 {
		return Choice{}, fmt.Errorf("invalid queue settings")
	}
	seconds, err := strconv.Atoi(fields[1])
	return Choice{seconds, fields[2]}, err
}
func Join(ctx context.Context, client *redis.Client, id, category string, choices ...Choice) error {
	return client.Set(ctx, presencePrefix+id, presence(category, choices), 20*time.Minute).Err()
}

func Active(ctx context.Context, client *redis.Client, id, category string, choices ...Choice) bool {
	v, err := client.Get(ctx, presencePrefix+id).Result()
	return err == nil && ((len(choices) == 0 && strings.Split(v, "|")[0] == category) || (len(choices) > 0 && v == presence(category, choices)))
}

func Leave(ctx context.Context, client *redis.Client, prefix, id string) error {
	return client.Eval(ctx, `redis.call('DEL',KEYS[1]); for i=2,#KEYS do redis.call('ZREM',KEYS[i],ARGV[1]) end; return 1`, keys(prefix, id), id).Err()
}

// A delayed Kafka join cannot reinsert a cancelled player. Rejoining a
// different category removes the previous category entry atomically.
func Enqueue(ctx context.Context, client *redis.Client, prefix, id, category string, choices ...Choice) error {
	return client.Eval(ctx, `if redis.call('GET',KEYS[1]) ~= ARGV[2] then return 0 end; for i=2,#KEYS do redis.call('ZREM',KEYS[i],ARGV[1]) end; redis.call('ZADD',ARGV[3],ARGV[4],ARGV[1]); return 1`, keys(prefix, id), id, presence(category, choices), prefix+":"+category, time.Now().UnixMicro()).Err()
}
