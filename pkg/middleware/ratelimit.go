package middleware

import (
	"net/http"
	"time"

	"api-core/config"
	"api-core/pkg/ratelimit"

	"github.com/go-redis/redis/v8"
)

// RateLimitMiddleware creates rate limiting middleware with default configuration
func RateLimitMiddleware(redisClient *redis.Client) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	// Use default configuration
	return ratelimit.RateLimitByUserOrIP(
		rateLimiter,
		rateLimitConfig.DefaultRule.Requests,
		rateLimitConfig.DefaultRule.Duration,
	)
}

// AuthRateLimitMiddleware creates rate limiting middleware for auth routes
func AuthRateLimitMiddleware(redisClient *redis.Client) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	// More restrictive rules for auth routes
	return ratelimit.RateLimitByIP(rateLimiter, 5, 15*60*time.Second) // 5 requests per 15 minutes
}

// UploadRateLimitMiddleware creates rate limiting middleware for upload routes
func UploadRateLimitMiddleware(redisClient *redis.Client) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	// More restrictive rules for upload routes
	return ratelimit.RateLimitByIP(rateLimiter, 10, 5*60*time.Second) // 10 requests per 5 minutes
}

// GlobalRateLimitMiddleware creates global rate limiting middleware
func GlobalRateLimitMiddleware(redisClient *redis.Client) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	// Global rate limit by IP
	return ratelimit.RateLimitByIP(rateLimiter, 1000, 60*60*time.Second) // 1000 requests per hour
}

// RateLimitByIP creates rate limiting middleware by IP
func RateLimitByIP(redisClient *redis.Client, requests int, duration time.Duration) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	return ratelimit.RateLimitByIP(rateLimiter, requests, duration*time.Second)
}

// RateLimitByUserOrIP creates rate limiting middleware by user ID if authenticated, otherwise IP
func RateLimitByUserOrIP(redisClient *redis.Client, requests int, duration time.Duration) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	return ratelimit.RateLimitByUserOrIP(rateLimiter, requests, duration*time.Second)
}

// RateLimitByIPAndRoute creates rate limiting middleware by IP and route
func RateLimitByIPAndRoute(redisClient *redis.Client, requests int, duration time.Duration) func(http.Handler) http.Handler {
	// Load rate limit configuration
	rateLimitConfig := config.LoadRateLimitConfig()

	if !rateLimitConfig.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create rate limiter
	rateLimiter := config.CreateRateLimiter(redisClient, rateLimitConfig)

	return ratelimit.RateLimitByIPAndRoute(rateLimiter, requests, duration*time.Second)
}
