package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Redis     *redis.Client
	KeyPrefix string
}

// RateLimitRule defines a rate limiting rule
type RateLimitRule struct {
	Requests int           // Number of requests allowed
	Duration time.Duration // Time window
	Key      string        // Key for Redis storage
}

// RateLimitResult holds the result of rate limiting check
type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	ResetTime  time.Time
	RetryAfter time.Duration
}

// RateLimiter handles rate limiting operations
type RateLimiter struct {
	config RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config: config,
	}
}

// CheckRateLimit checks if a request is allowed based on the rule
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, rule RateLimitRule) (*RateLimitResult, error) {
	key := rl.config.KeyPrefix + ":" + rule.Key

	// Use Redis pipeline for atomic operations
	pipe := rl.config.Redis.Pipeline()

	// Increment counter
	incr := pipe.Incr(ctx, key)

	// Set expiration only if key doesn't exist (NX flag)
	pipe.ExpireNX(ctx, key, rule.Duration)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	current, err := incr.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get current count: %w", err)
	}

	// Check if rate limit exceeded
	if current > int64(rule.Requests) {
		// Get TTL to calculate retry after
		ttl, err := rl.config.Redis.TTL(ctx, key).Result()
		if err != nil {
			ttl = rule.Duration // Fallback to rule duration
		}

		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now().Add(ttl),
			RetryAfter: ttl,
		}, nil
	}

	// Calculate remaining requests
	remaining := rule.Requests - int(current)
	if remaining < 0 {
		remaining = 0
	}

	// Get TTL for reset time
	ttl, err := rl.config.Redis.TTL(ctx, key).Result()
	if err != nil {
		ttl = rule.Duration // Fallback to rule duration
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		ResetTime:  time.Now().Add(ttl),
		RetryAfter: 0,
	}, nil
}

// GetRateLimitInfo gets current rate limit information without incrementing
func (rl *RateLimiter) GetRateLimitInfo(ctx context.Context, rule RateLimitRule) (*RateLimitResult, error) {
	key := rl.config.KeyPrefix + ":" + rule.Key

	// Get current count
	current, err := rl.config.Redis.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist, so no requests made yet
			return &RateLimitResult{
				Allowed:    true,
				Remaining:  rule.Requests,
				ResetTime:  time.Now().Add(rule.Duration),
				RetryAfter: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get current count: %w", err)
	}

	// Check if rate limit exceeded
	if current >= rule.Requests {
		// Get TTL to calculate retry after
		ttl, err := rl.config.Redis.TTL(ctx, key).Result()
		if err != nil {
			ttl = rule.Duration // Fallback to rule duration
		}

		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now().Add(ttl),
			RetryAfter: ttl,
		}, nil
	}

	// Calculate remaining requests
	remaining := rule.Requests - current
	if remaining < 0 {
		remaining = 0
	}

	// Get TTL for reset time
	ttl, err := rl.config.Redis.TTL(ctx, key).Result()
	if err != nil {
		ttl = rule.Duration // Fallback to rule duration
	}

	return &RateLimitResult{
		Allowed:    true,
		Remaining:  remaining,
		ResetTime:  time.Now().Add(ttl),
		RetryAfter: 0,
	}, nil
}

// ResetRateLimit resets the rate limit for a specific key
func (rl *RateLimiter) ResetRateLimit(ctx context.Context, key string) error {
	fullKey := rl.config.KeyPrefix + ":" + key
	return rl.config.Redis.Del(ctx, fullKey).Err()
}
