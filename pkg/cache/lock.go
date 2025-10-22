package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Distributed Lock Implementation

var (
	ErrLockFailed  = errors.New("failed to acquire lock")
	ErrLockNotHeld = errors.New("lock not held")
	ErrLockTimeout = errors.New("lock acquisition timeout")
)

// Lock cố gắng acquire lock
// Returns true if lock acquired, false if already locked
func (c *redisCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Generate unique lock value
	lockValue := generateLockValue()

	// Try to set lock với NX (only if not exists)
	lockKey := "lock:" + key
	success, err := c.client.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return success, nil
}

// Unlock giải phóng lock
func (c *redisCache) Unlock(ctx context.Context, key string) error {
	lockKey := "lock:" + key
	return c.client.Del(ctx, lockKey).Err()
}

// LockAndWait cố gắng acquire lock và đợi nếu bị lock
// maxWait: thời gian đợi tối đa
func (c *redisCache) LockAndWait(ctx context.Context, key string, ttl time.Duration, maxWait time.Duration) (bool, error) {
	lockKey := "lock:" + key
	lockValue := generateLockValue()

	deadline := time.Now().Add(maxWait)
	retryInterval := 50 * time.Millisecond

	for {
		// Try acquire lock
		success, err := c.client.SetNX(ctx, lockKey, lockValue, ttl).Result()
		if err != nil {
			return false, fmt.Errorf("failed to acquire lock: %w", err)
		}

		if success {
			return true, nil
		}

		// Check if timeout
		if time.Now().After(deadline) {
			return false, ErrLockTimeout
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryInterval):
			// Increase retry interval (exponential backoff)
			retryInterval = min(retryInterval*2, 1*time.Second)
		}
	}
}

// UnlockWithValue giải phóng lock với value check (safer)
func (c *redisCache) UnlockWithValue(ctx context.Context, key string, value string) error {
	lockKey := "lock:" + key

	// Lua script để check value before delete (atomic)
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := c.client.Eval(ctx, script, []string{lockKey}, value).Result()
	if err != nil {
		return err
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// generateLockValue tạo unique value cho lock
func generateLockValue() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
