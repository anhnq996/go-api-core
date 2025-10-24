package config

import (
	"anhnq/api-core/pkg/cache"
	"fmt"

	"anhnq/api-core/pkg/utils"
)

// CacheConfig cấu hình cache
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// GetDefaultCacheConfig trả về config mặc định từ env
func GetDefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Host:     utils.GetEnv("REDIS_HOST", "localhost"),
		Port:     utils.GetEnv("REDIS_PORT", "6379"),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       utils.GetEnvInt("REDIS_DB", 0),
		PoolSize: utils.GetEnvInt("REDIS_POOL_SIZE", 10),
	}
}

// ConnectCache kết nối đến Redis
func ConnectCache(cfg CacheConfig) (cache.Cache, error) {
	cacheClient, err := cache.NewRedisCache(cache.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to cache: %w", err)
	}

	return cacheClient, nil
}
