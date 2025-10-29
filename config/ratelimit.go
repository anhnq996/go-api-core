package config

import (
	"time"

	"api-core/pkg/ratelimit"
	"api-core/pkg/utils"

	"github.com/go-redis/redis/v8"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool
	KeyPrefix   string
	DefaultRule RateLimitRule
	RouteRules  map[string]RateLimitRule
	IPRules     map[string]RateLimitRule
}

// RateLimitRule holds rate limiting rule configuration
type RateLimitRule struct {
	Requests int
	Duration time.Duration
}

// LoadRateLimitConfig loads rate limiting configuration from environment variables
func LoadRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:   utils.GetEnvBool("RATE_LIMIT_ENABLED", true),
		KeyPrefix: utils.GetEnv("RATE_LIMIT_KEY_PREFIX", "ratelimit"),
		DefaultRule: RateLimitRule{
			Requests: utils.GetEnvInt("RATE_LIMIT_DEFAULT_REQUESTS", 100),
			Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_DEFAULT_DURATION_MINUTES", 1)) * time.Minute,
		},
		RouteRules: getRouteRules(),
		IPRules:    getIPRules(),
	}
}

// getRouteRules returns route-specific rate limiting rules
func getRouteRules() map[string]RateLimitRule {
	rules := make(map[string]RateLimitRule)

	// Auth routes - more restrictive
	rules["/api/v1/auth/login"] = RateLimitRule{
		Requests: utils.GetEnvInt("RATE_LIMIT_AUTH_LOGIN_REQUESTS", 5),
		Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_AUTH_LOGIN_DURATION_MINUTES", 15)) * time.Minute,
	}

	rules["/api/v1/auth/register"] = RateLimitRule{
		Requests: utils.GetEnvInt("RATE_LIMIT_AUTH_REGISTER_REQUESTS", 3),
		Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_AUTH_REGISTER_DURATION_MINUTES", 60)) * time.Minute,
	}

	rules["/api/v1/auth/refresh"] = RateLimitRule{
		Requests: utils.GetEnvInt("RATE_LIMIT_AUTH_REFRESH_REQUESTS", 10),
		Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_AUTH_REFRESH_DURATION_MINUTES", 5)) * time.Minute,
	}

	// User routes
	rules["/api/v1/users"] = RateLimitRule{
		Requests: utils.GetEnvInt("RATE_LIMIT_USERS_REQUESTS", 50),
		Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_USERS_DURATION_MINUTES", 1)) * time.Minute,
	}

	// Upload routes - more restrictive
	rules["/api/v1/users/*/avatar"] = RateLimitRule{
		Requests: utils.GetEnvInt("RATE_LIMIT_UPLOAD_REQUESTS", 10),
		Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_UPLOAD_DURATION_MINUTES", 5)) * time.Minute,
	}

	return rules
}

// getIPRules returns IP-specific rate limiting rules
func getIPRules() map[string]RateLimitRule {
	rules := make(map[string]RateLimitRule)

	// Global IP rate limit
	rules["global"] = RateLimitRule{
		Requests: utils.GetEnvInt("RATE_LIMIT_IP_GLOBAL_REQUESTS", 1000),
		Duration: time.Duration(utils.GetEnvInt("RATE_LIMIT_IP_GLOBAL_DURATION_MINUTES", 60)) * time.Minute,
	}

	return rules
}

// CreateRateLimiter creates a rate limiter instance
func CreateRateLimiter(redisClient *redis.Client, config *RateLimitConfig) *ratelimit.RateLimiter {
	return ratelimit.NewRateLimiter(ratelimit.RateLimitConfig{
		Redis:     redisClient,
		KeyPrefix: config.KeyPrefix,
	})
}
