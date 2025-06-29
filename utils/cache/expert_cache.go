package cache

import (
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"context"
	"fmt"
	"time"
)

const expertDetailTTL = 6 * time.Hour

type ExpertCache interface {
	GetExpertDetail(ctx context.Context, expertID string) (*dtoexperts.ExpertFullDetailResponse, error)
	SetExpertDetail(ctx context.Context, expertID string, expert *dtoexperts.ExpertFullDetailResponse) error
	DeleteExpertDetail(ctx context.Context, expertID string) error
}

type RedisExpertCache struct {
	redis *RedisCache
}

func NewRedisExpertCache(redis *RedisCache) *RedisExpertCache {
	return &RedisExpertCache{redis: redis}
}

func (c *RedisExpertCache) GetExpertDetail(ctx context.Context, expertID string) (*dtoexperts.ExpertFullDetailResponse, error) {
	key := fmt.Sprintf("expert:detail:%s", expertID)
	var expert dtoexperts.ExpertFullDetailResponse
	err := c.redis.Get(ctx, key, &expert)
	if err != nil {
		return nil, err
	}
	return &expert, nil
}

func (c *RedisExpertCache) SetExpertDetail(ctx context.Context, expertID string, expert *dtoexperts.ExpertFullDetailResponse) error {
	key := fmt.Sprintf("expert:detail:%s", expertID)
	return c.redis.Set(ctx, key, expert, expertDetailTTL)
}

func (c *RedisExpertCache) DeleteExpertDetail(ctx context.Context, expertID string) error {
	key := fmt.Sprintf("expert:detail:%s", expertID)
	return c.redis.Delete(ctx, key)
}
