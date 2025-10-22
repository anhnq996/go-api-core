package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config cấu hình Redis
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// NewRedisCache tạo Redis cache instance
func NewRedisCache(cfg Config) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisCache{client: client}, nil
}

// Basic Operations

// Get lấy giá trị từ key
func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set lưu giá trị với TTL
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Convert value to string
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		// JSON encode for complex types
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		strValue = string(data)
	}

	return c.client.Set(ctx, key, strValue, ttl).Err()
}

// Del xóa keys
func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists kiểm tra key tồn tại
func (c *redisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire set TTL cho key
func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

// TTL lấy thời gian còn lại của key
func (c *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Remember pattern - Get from cache or execute callback
func (c *redisCache) Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Try get from cache
	val, err := c.Get(ctx, key)
	if err == nil {
		// Cache hit - decode JSON
		var result interface{}
		if err := json.Unmarshal([]byte(val), &result); err != nil {
			// If not JSON, return as string
			return val, nil
		}
		return result, nil
	}

	// Cache miss - execute callback
	if err != redis.Nil {
		return nil, fmt.Errorf("cache error: %w", err)
	}

	result, err := callback()
	if err != nil {
		return nil, err
	}

	// Save to cache
	if err := c.Set(ctx, key, result, ttl); err != nil {
		// Don't fail if cache set fails - just log and return result
		fmt.Printf("Warning: failed to set cache: %v\n", err)
	}

	return result, nil
}
