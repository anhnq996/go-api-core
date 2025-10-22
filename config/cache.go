package config

import (
	"anhnq/api-core/pkg/cache"
	"fmt"
	"strconv"
)

// CacheConfig cấu hình cache
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// GetDefaultCacheConfig trả về config mặc định
func GetDefaultCacheConfig() CacheConfig {
	// Use getEnv from database.go package
	dbNum, _ := strconv.Atoi("0")
	poolSize, _ := strconv.Atoi("10")

	return CacheConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       dbNum,
		PoolSize: poolSize,
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
