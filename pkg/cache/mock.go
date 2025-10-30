package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// MockCache implements cache.Cache interface for testing
type MockCache struct {
	data map[string]interface{}
	ttl  map[string]time.Time
	mu   sync.RWMutex
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

// Get retrieves a value from cache
func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if key exists and is not expired
	if ttl, exists := m.ttl[key]; exists {
		if time.Now().After(ttl) {
			// Key expired, remove it
			delete(m.data, key)
			delete(m.ttl, key)
			return "", ErrCacheMiss
		}
	}

	if value, exists := m.data[key]; exists {
		// Convert to string
		if str, ok := value.(string); ok {
			return str, nil
		}
		// If not string, convert to JSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}
	return "", ErrCacheMiss
}

// Set stores a value in cache
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	if ttl > 0 {
		m.ttl[key] = time.Now().Add(ttl)
	}
	return nil
}

// Del removes values from cache
func (m *MockCache) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		delete(m.data, key)
		delete(m.ttl, key)
	}
	return nil
}

// Exists checks if keys exist in cache
func (m *MockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := int64(0)
	for _, key := range keys {
		if ttl, exists := m.ttl[key]; exists {
			if time.Now().After(ttl) {
				// Key expired, remove it
				delete(m.data, key)
				delete(m.ttl, key)
				continue
			}
		}

		if _, exists := m.data[key]; exists {
			count++
		}
	}
	return count, nil
}

// Expire sets expiration for a key
func (m *MockCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; exists {
		m.ttl[key] = time.Now().Add(ttl)
		return nil
	}
	return ErrCacheMiss
}

// TTL returns the time to live for a key
func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ttl, exists := m.ttl[key]; exists {
		if time.Now().After(ttl) {
			// Key expired
			return -1, ErrCacheMiss
		}
		return time.Until(ttl), nil
	}
	return -2, ErrCacheMiss // Key doesn't exist
}

// Remember executes a function and caches the result
func (m *MockCache) Remember(ctx context.Context, key string, ttl time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, err := m.Get(ctx, key); err == nil {
		return value, nil
	}

	// Execute callback
	result, err := callback()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := m.Set(ctx, key, result, ttl); err != nil {
		return result, err // Return result even if caching fails
	}

	return result, nil
}

// Hash operations - simplified implementations
func (m *MockCache) HSet(ctx context.Context, key, field string, value interface{}) error {
	return m.Set(ctx, key+":"+field, value, 0)
}

func (m *MockCache) HGet(ctx context.Context, key, field string) (string, error) {
	return m.Get(ctx, key+":"+field)
}

func (m *MockCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	// Simplified implementation - return empty map
	return make(map[string]string), nil
}

func (m *MockCache) HDel(ctx context.Context, key string, fields ...string) error {
	for _, field := range fields {
		m.Del(ctx, key+":"+field)
	}
	return nil
}

func (m *MockCache) HExists(ctx context.Context, key, field string) (bool, error) {
	_, err := m.Get(ctx, key+":"+field)
	return err == nil, nil
}

// Set operations - simplified implementations
func (m *MockCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return nil // Simplified
}

func (m *MockCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return nil // Simplified
}

func (m *MockCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil // Simplified
}

func (m *MockCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return false, nil // Simplified
}

func (m *MockCache) SCard(ctx context.Context, key string) (int64, error) {
	return 0, nil // Simplified
}

// List operations - simplified implementations
func (m *MockCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return nil // Simplified
}

func (m *MockCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return nil // Simplified
}

func (m *MockCache) LPop(ctx context.Context, key string) (string, error) {
	return "", ErrCacheMiss // Simplified
}

func (m *MockCache) RPop(ctx context.Context, key string) (string, error) {
	return "", ErrCacheMiss // Simplified
}

func (m *MockCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return []string{}, nil // Simplified
}

// Distributed lock - simplified implementations
func (m *MockCache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return true, nil // Simplified - always succeeds
}

func (m *MockCache) Unlock(ctx context.Context, key string) error {
	return nil // Simplified
}

func (m *MockCache) LockAndWait(ctx context.Context, key string, ttl time.Duration, maxWait time.Duration) (bool, error) {
	return true, nil // Simplified - always succeeds
}

// Utility methods
func (m *MockCache) Ping(ctx context.Context) error {
	return nil
}

func (m *MockCache) FlushDB(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]interface{})
	m.ttl = make(map[string]time.Time)
	return nil
}

func (m *MockCache) GetRedisClient() *redis.Client {
	return nil
}

func (m *MockCache) Close() error {
	return nil
}
