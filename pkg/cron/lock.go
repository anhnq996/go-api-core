package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisLockManager implements LockManager using Redis
type RedisLockManager struct {
	client *redis.Client
	prefix string
}

// NewRedisLockManager creates a new Redis lock manager
func NewRedisLockManager(client *redis.Client, prefix string) *RedisLockManager {
	if prefix == "" {
		prefix = "cron:lock:"
	}
	return &RedisLockManager{
		client: client,
		prefix: prefix,
	}
}

// AcquireLock attempts to acquire a lock for the given job
func (r *RedisLockManager) AcquireLock(ctx context.Context, jobName string, ttl time.Duration) (bool, error) {
	key := r.getLockKey(jobName)

	// Try to set the lock with NX (only if not exists) and EX (expiration)
	result := r.client.SetNX(ctx, key, time.Now().Unix(), ttl)
	if result.Err() != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", result.Err())
	}

	return result.Val(), nil
}

// ReleaseLock releases the lock for the given job
func (r *RedisLockManager) ReleaseLock(ctx context.Context, jobName string) error {
	key := r.getLockKey(jobName)

	result := r.client.Del(ctx, key)
	if result.Err() != nil {
		return fmt.Errorf("failed to release lock: %w", result.Err())
	}

	return nil
}

// ExtendLock extends the lock TTL for the given job
func (r *RedisLockManager) ExtendLock(ctx context.Context, jobName string, ttl time.Duration) error {
	key := r.getLockKey(jobName)

	// Check if lock exists and extend it
	result := r.client.Expire(ctx, key, ttl)
	if result.Err() != nil {
		return fmt.Errorf("failed to extend lock: %w", result.Err())
	}

	if !result.Val() {
		return fmt.Errorf("lock does not exist or has expired")
	}

	return nil
}

// IsLocked checks if a job is currently locked
func (r *RedisLockManager) IsLocked(ctx context.Context, jobName string) (bool, error) {
	key := r.getLockKey(jobName)

	result := r.client.Exists(ctx, key)
	if result.Err() != nil {
		return false, fmt.Errorf("failed to check lock: %w", result.Err())
	}

	return result.Val() > 0, nil
}

// getLockKey returns the Redis key for the lock
func (r *RedisLockManager) getLockKey(jobName string) string {
	return r.prefix + jobName
}

// MemoryLockManager implements LockManager using in-memory locks (for single instance)
type MemoryLockManager struct {
	locks map[string]*memoryLock
}

type memoryLock struct {
	acquiredAt time.Time
	ttl        time.Duration
}

// NewMemoryLockManager creates a new in-memory lock manager
func NewMemoryLockManager() *MemoryLockManager {
	return &MemoryLockManager{
		locks: make(map[string]*memoryLock),
	}
}

// AcquireLock attempts to acquire a lock for the given job
func (m *MemoryLockManager) AcquireLock(ctx context.Context, jobName string, ttl time.Duration) (bool, error) {
	// Check if lock exists and is still valid
	if lock, exists := m.locks[jobName]; exists {
		if time.Since(lock.acquiredAt) < lock.ttl {
			return false, nil // Lock is still valid
		}
		// Lock has expired, remove it
		delete(m.locks, jobName)
	}

	// Acquire new lock
	m.locks[jobName] = &memoryLock{
		acquiredAt: time.Now(),
		ttl:        ttl,
	}

	return true, nil
}

// ReleaseLock releases the lock for the given job
func (m *MemoryLockManager) ReleaseLock(ctx context.Context, jobName string) error {
	delete(m.locks, jobName)
	return nil
}

// ExtendLock extends the lock TTL for the given job
func (m *MemoryLockManager) ExtendLock(ctx context.Context, jobName string, ttl time.Duration) error {
	if lock, exists := m.locks[jobName]; exists {
		lock.ttl = ttl
		return nil
	}
	return fmt.Errorf("lock does not exist")
}

// IsLocked checks if a job is currently locked
func (m *MemoryLockManager) IsLocked(ctx context.Context, jobName string) (bool, error) {
	if lock, exists := m.locks[jobName]; exists {
		if time.Since(lock.acquiredAt) < lock.ttl {
			return true, nil
		}
		// Lock has expired, remove it
		delete(m.locks, jobName)
	}
	return false, nil
}

// CleanupExpiredLocks removes expired locks from memory
func (m *MemoryLockManager) CleanupExpiredLocks() {
	now := time.Now()
	for jobName, lock := range m.locks {
		if now.Sub(lock.acquiredAt) >= lock.ttl {
			delete(m.locks, jobName)
		}
	}
}
