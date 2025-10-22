package cache

import (
	"context"
	"errors"
	"time"
)

// noopCache implements Cache interface but does nothing (no-op)
// Used when Redis is not available
type noopCache struct{}

// NewNoopCache tạo no-op cache (fallback khi Redis không có)
func NewNoopCache() Cache {
	return &noopCache{}
}

var ErrCacheNotAvailable = errors.New("cache not available")

// Basic operations - always miss
func (c *noopCache) Get(ctx context.Context, key string) (string, error) {
	return "", ErrCacheNotAvailable
}

func (c *noopCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil // No-op
}

func (c *noopCache) Del(ctx context.Context, keys ...string) error {
	return nil // No-op
}

func (c *noopCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 0, nil
}

func (c *noopCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (c *noopCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (c *noopCache) Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Always execute callback (no caching)
	return callback()
}

// Hash operations
func (c *noopCache) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return nil
}

func (c *noopCache) HGet(ctx context.Context, key string, field string) (string, error) {
	return "", ErrCacheNotAvailable
}

func (c *noopCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, ErrCacheNotAvailable
}

func (c *noopCache) HDel(ctx context.Context, key string, fields ...string) error {
	return nil
}

func (c *noopCache) HExists(ctx context.Context, key string, field string) (bool, error) {
	return false, nil
}

// Set operations
func (c *noopCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

func (c *noopCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

func (c *noopCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return nil, nil
}

func (c *noopCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return false, nil
}

func (c *noopCache) SCard(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

// List operations
func (c *noopCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

func (c *noopCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

func (c *noopCache) LPop(ctx context.Context, key string) (string, error) {
	return "", ErrCacheNotAvailable
}

func (c *noopCache) RPop(ctx context.Context, key string) (string, error) {
	return "", ErrCacheNotAvailable
}

func (c *noopCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return nil, nil
}

// Lock operations - always succeed (no locking)
func (c *noopCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return true, nil // Always allow
}

func (c *noopCache) Unlock(ctx context.Context, key string) error {
	return nil
}

func (c *noopCache) LockAndWait(ctx context.Context, key string, ttl time.Duration, maxWait time.Duration) (bool, error) {
	return true, nil // Always allow
}

// Utility
func (c *noopCache) Ping(ctx context.Context) error {
	return ErrCacheNotAvailable
}

func (c *noopCache) FlushDB(ctx context.Context) error {
	return nil
}

func (c *noopCache) Close() error {
	return nil
}
