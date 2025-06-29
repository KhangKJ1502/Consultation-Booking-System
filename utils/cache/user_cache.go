package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrCacheUnavailable = errors.New("cache service unavailable")
	ErrInvalidToken     = errors.New("invalid token")
)

type UserCache interface {
	GetTokenFromCache(ctx context.Context, key string) (uuid.UUID, error)
	IsRedisHealthy(ctx context.Context) bool
	InvalidateToken(ctx context.Context, token string) error
	CheckTokenExists(ctx context.Context, key string) (bool, error)
	SetToken(ctx context.Context, token string, userID uuid.UUID, expiration time.Duration) error
}

type RedisUserCache struct {
	redisCache *RedisCache
	logger     *zap.Logger
}

// NewRedisUserCache creates a new Redis user cache instance
func NewRedisUserCache(redisCache *RedisCache, logger *zap.Logger) *RedisUserCache {
	return &RedisUserCache{
		redisCache: redisCache,
		logger:     logger,
	}
}

// GetTokenFromCache retrieves user ID from Redis cache using token key
func (s *RedisUserCache) GetTokenFromCache(ctx context.Context, key string) (uuid.UUID, error) {
	if s.redisCache == nil {
		s.logger.Error("Redis cache not initialized")
		return uuid.Nil, ErrCacheUnavailable
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var userID uuid.UUID

	// Get data from Redis
	err := s.redisCache.Get(timeoutCtx, key, &userID)
	if err != nil {
		// Check if it's a timeout error
		if timeoutCtx.Err() != nil {
			s.logger.Warn("Redis timeout when getting token",
				zap.Error(timeoutCtx.Err()),
				zap.String("key", key))
			return uuid.Nil, fmt.Errorf("redis timeout: %w", timeoutCtx.Err())
		}

		// Check if it's a "key not found" error
		if errors.Is(err, redis.Nil) {
			s.logger.Debug("Token not found in cache", zap.String("key", key))
			return uuid.Nil, redis.Nil
		}

		s.logger.Error("Redis get error",
			zap.Error(err),
			zap.String("key", key))
		return uuid.Nil, fmt.Errorf("redis get error: %w", err)
	}

	s.logger.Debug("Token retrieved successfully",
		zap.String("key", key),
		zap.String("userID", userID.String()))
	return userID, nil
}

// CheckTokenExists checks if token exists in Redis with error handling
func (s *RedisUserCache) CheckTokenExists(ctx context.Context, key string) (bool, error) {
	if s.redisCache == nil {
		s.logger.Error("Redis cache not initialized")
		return false, ErrCacheUnavailable
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check existence using EXISTS command
	exists, err := s.redisCache.Client.Exists(timeoutCtx, key).Result()
	if err != nil {
		// Check if context was cancelled due to timeout
		if timeoutCtx.Err() != nil {
			s.logger.Warn("Redis timeout when checking token existence",
				zap.Error(timeoutCtx.Err()),
				zap.String("key", key))
			return false, fmt.Errorf("redis timeout: %w", timeoutCtx.Err())
		}

		s.logger.Error("Redis exists check error",
			zap.Error(err),
			zap.String("key", key))
		return false, fmt.Errorf("redis exists error: %w", err)
	}

	return exists > 0, nil
}

// IsRedisHealthy checks if Redis connection is healthy
func (s *RedisUserCache) IsRedisHealthy(ctx context.Context) bool {
	if s.redisCache == nil {
		s.logger.Warn("Redis cache not initialized")
		return false
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Try a simple ping operation
	err := s.redisCache.Client.Ping(timeoutCtx).Err()
	if err != nil {
		s.logger.Warn("Redis health check failed", zap.Error(err))
		return false
	}

	return true
}

// InvalidateToken manually invalidates a token by removing it from cache
func (s *RedisUserCache) InvalidateToken(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token must not be empty")
	}

	if s.redisCache == nil {
		s.logger.Error("Redis cache not initialized")
		return ErrCacheUnavailable
	}

	key := "auth:" + token

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.redisCache.Delete(timeoutCtx, key)
	if err != nil {
		s.logger.Error("Failed to invalidate token",
			zap.Error(err),
			zap.String("key", key))
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	s.logger.Info("Token invalidated successfully", zap.String("key", key))
	return nil
}

// SetToken stores a token with userID in Redis cache
func (s *RedisUserCache) SetToken(ctx context.Context, token string, userID uuid.UUID, expiration time.Duration) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token must not be empty")
	}

	if userID == uuid.Nil {
		return fmt.Errorf("userID must not be empty")
	}

	if s.redisCache == nil {
		s.logger.Error("Redis cache not initialized")
		return ErrCacheUnavailable
	}

	key := "auth:" + token

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.redisCache.Set(timeoutCtx, key, userID, expiration)
	if err != nil {
		s.logger.Error("Failed to set token in cache",
			zap.Error(err),
			zap.String("key", key),
			zap.String("userID", userID.String()))
		return fmt.Errorf("failed to set token: %w", err)
	}

	s.logger.Info("Token stored successfully",
		zap.String("key", key),
		zap.String("userID", userID.String()),
		zap.Duration("expiration", expiration))
	return nil
}
