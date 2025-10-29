package ratelimit

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-core/pkg/exception"

	"github.com/go-chi/chi/v5"
)

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	RateLimiter *RateLimiter
	Rules       map[string]RateLimitRule // Route pattern -> rule mapping
	DefaultRule RateLimitRule            // Default rule for unmatched routes
	KeyFunc     KeyFunc                  // Function to generate rate limit key
}

// KeyFunc defines a function to generate rate limit keys
type KeyFunc func(r *http.Request) string

// Default key functions
var (
	// KeyByIP generates key based on client IP
	KeyByIP = func(r *http.Request) string {
		// Get real IP from headers (for proxies)
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			return "ip:" + ip
		}
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			// Take first IP if multiple
			ips := strings.Split(ip, ",")
			return "ip:" + strings.TrimSpace(ips[0])
		}
		// Fallback to remote address
		ip := strings.Split(r.RemoteAddr, ":")[0]
		return "ip:" + ip
	}

	// KeyByUserID generates key based on user ID (requires auth middleware)
	KeyByUserID = func(r *http.Request) string {
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			return "user:" + userID
		}
		return "anonymous"
	}

	// KeyByIPAndRoute generates key based on IP and route
	KeyByIPAndRoute = func(r *http.Request) string {
		ipKey := KeyByIP(r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		return ipKey + ":route:" + route
	}

	// KeyByUserOrIP generates key based on user ID if authenticated, otherwise IP
	KeyByUserOrIP = func(r *http.Request) string {
		// Check if user is authenticated (has user ID in header or context)
		// if userID := r.Header.Get("X-User-ID"); userID != "" {
		// 	return "user:" + userID
		// }

		// Check if user ID is in context (from JWT middleware)
		if userID := r.Context().Value("user_id"); userID != nil {
			if id, ok := userID.(string); ok && id != "" {
				return "user:" + id
			}
		}

		// Fallback to IP
		return KeyByIP(r)
	}
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get route pattern
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}

			// Find matching rule
			rule := config.DefaultRule
			if configRule, exists := config.Rules[route]; exists {
				rule = configRule
			}

			// Generate rate limit key
			key := config.KeyFunc(r)
			rule.Key = key

			// Check rate limit
			result, err := config.RateLimiter.CheckRateLimit(r.Context(), rule)
			if err != nil {
				// Log error but allow request to proceed
				http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rule.Requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

			if !result.Allowed {
				w.Header().Set("X-RateLimit-Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				panic(exception.Exception{
					Message: "Rate limit exceeded",
					Code:    "RATE_LIMIT_EXCEEDED",
				})
				// 	http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				// 	return
			}

			// Proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitWithConfig creates a rate limiting middleware with custom configuration
func RateLimitWithConfig(rateLimiter *RateLimiter, requests int, duration time.Duration, keyFunc KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate rate limit key
			key := keyFunc(r)

			// Create rule
			rule := RateLimitRule{
				Requests: requests,
				Duration: duration,
				Key:      key,
			}

			// Check rate limit
			result, err := rateLimiter.CheckRateLimit(r.Context(), rule)
			if err != nil {
				// Log error but allow request to proceed
				http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rule.Requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

			if !result.Allowed {
				w.Header().Set("X-RateLimit-Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByIP creates a simple rate limiting middleware by IP
func RateLimitByIP(rateLimiter *RateLimiter, requests int, duration time.Duration) func(http.Handler) http.Handler {
	config := MiddlewareConfig{
		RateLimiter: rateLimiter,
		DefaultRule: RateLimitRule{
			Requests: requests,
			Duration: duration,
		},
		KeyFunc: KeyByIP,
	}

	return RateLimitMiddleware(config)
}

// RateLimitByRoute creates a rate limiting middleware by route
func RateLimitByRoute(rateLimiter *RateLimiter, rules map[string]RateLimitRule) func(http.Handler) http.Handler {
	config := MiddlewareConfig{
		RateLimiter: rateLimiter,
		Rules:       rules,
		DefaultRule: RateLimitRule{
			Requests: 100,
			Duration: time.Minute,
		},
		KeyFunc: KeyByIPAndRoute,
	}

	return RateLimitMiddleware(config)
}

// RateLimitByUser creates a rate limiting middleware by user ID
func RateLimitByUser(rateLimiter *RateLimiter, requests int, duration time.Duration) func(http.Handler) http.Handler {
	config := MiddlewareConfig{
		RateLimiter: rateLimiter,
		DefaultRule: RateLimitRule{
			Requests: requests,
			Duration: duration,
		},
		KeyFunc: KeyByUserID,
	}

	return RateLimitMiddleware(config)
}

// RateLimitByUserOrIP creates a rate limiting middleware by user ID if authenticated, otherwise IP
func RateLimitByUserOrIP(rateLimiter *RateLimiter, requests int, duration time.Duration) func(http.Handler) http.Handler {
	return RateLimitWithConfig(rateLimiter, requests, duration, KeyByUserOrIP)
}

// RateLimitByIPAndRoute creates a rate limiting middleware by IP and route
func RateLimitByIPAndRoute(rateLimiter *RateLimiter, requests int, duration time.Duration) func(http.Handler) http.Handler {
	return RateLimitWithConfig(rateLimiter, requests, duration, KeyByIPAndRoute)
}
