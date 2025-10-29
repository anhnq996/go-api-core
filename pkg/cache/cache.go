package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache interface định nghĩa các operations
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Remember pattern
	Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error)

	// Hash operations
	HSet(ctx context.Context, key string, field string, value interface{}) error
	HGet(ctx context.Context, key string, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HExists(ctx context.Context, key string, field string) (bool, error)

	// Redis client access for rate limiting
	GetRedisClient() *redis.Client

	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)

	// List operations
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// Distributed lock
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
	LockAndWait(ctx context.Context, key string, ttl time.Duration, maxWait time.Duration) (bool, error)

	// Utility
	Ping(ctx context.Context) error
	FlushDB(ctx context.Context) error
	Close() error
}

// redisCache implements Cache interface
type redisCache struct {
	client *redis.Client
}

var _ Cache = (*redisCache)(nil)
