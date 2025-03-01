package cachemanager

import (
	"BrainBlitz.com/game/adapter/redis"
	"context"
	"fmt"
	"time"
)

type CacheManager struct {
	cache *redis.Adapter
}

func NewCacheManager(cache *redis.Adapter) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

func (c *CacheManager) Set(ctx context.Context, key string, value any, expire time.Duration) error {

	err := c.cache.Client().Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheManager) Get(ctx context.Context, key string) (string, error) {
	data, err := c.cache.Client().Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return data, nil
}

func (c *CacheManager) Delete(ctx context.Context, keys ...string) error {

	if err := c.cache.Client().Del(ctx, keys...).Err(); err != nil {
		return err
	}

	return nil
}

func (c *CacheManager) GetTTL(ctx context.Context, key string) (int64, bool, error) {

	ttl, err := c.cache.Client().TTL(ctx, key).Result()
	if err != nil {
		return 0, false, err
	}

	// -2 if the key does not exist
	if ttl == -2*time.Nanosecond {
		return 0, false, fmt.Errorf("key does not exist")
	}

	return int64(ttl.Seconds()), true, nil
}
