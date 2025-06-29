package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{Client: client}
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rc.Client.Set(ctx, key, data, expiration).Err()
}

func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.Client.Del(ctx, key).Err()
}
