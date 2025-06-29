package initialize

import (
	"cbs_backend/global"
	"cbs_backend/utils/cache"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *cache.RedisCache {
	cfg := global.ConfigConection.RedisCF

	if cfg.Host == "" || cfg.Port == "" {
		log.Println("⚠️ Redis configuration not provided, skipping Redis initialization")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	global.Redis = rdb
	fmt.Printf("✅ Redis connected: %s:%s\n", cfg.Host, cfg.Port)
	return cache.NewRedisCache(rdb)
}
