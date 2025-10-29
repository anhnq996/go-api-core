# Rate Limiting Package

Package này cung cấp rate limiting functionality sử dụng Redis để lưu trữ và quản lý rate limits.

## Features

- **Redis-based**: Sử dụng Redis để lưu trữ rate limit counters
- **Configurable**: Có thể cấu hình theo route, IP, hoặc user
- **Flexible**: Hỗ trợ nhiều loại rate limiting rules
- **Middleware**: Tích hợp dễ dàng với chi router
- **Headers**: Tự động thêm rate limit headers vào response

## Usage

### 1. Basic Rate Limiting

```go
import "api-core/pkg/ratelimit"

// Create rate limiter
rateLimiter := ratelimit.NewRateLimiter(ratelimit.RateLimitConfig{
    Redis:     redisClient,
    KeyPrefix: "ratelimit",
})

// Check rate limit
rule := ratelimit.RateLimitRule{
    Requests: 100,
    Duration: time.Minute,
    Key:      "ip:192.168.1.1",
}

result, err := rateLimiter.CheckRateLimit(ctx, rule)
if err != nil {
    // Handle error
}

if !result.Allowed {
    // Rate limit exceeded
    fmt.Printf("Rate limit exceeded. Retry after: %v", result.RetryAfter)
}
```

### 2. Middleware Usage

```go
import "api-core/pkg/ratelimit"

// Rate limiting by IP
r.Use(ratelimit.RateLimitByIP(rateLimiter, 100, time.Minute))

// Rate limiting by user ID if authenticated, otherwise IP
r.Use(ratelimit.RateLimitByUserOrIP(rateLimiter, 100, time.Minute))

// Rate limiting by IP and route
r.Use(ratelimit.RateLimitByIPAndRoute(rateLimiter, 50, time.Minute))

// Rate limiting by user ID
r.Use(ratelimit.RateLimitByUser(rateLimiter, 1000, time.Hour))
```

### 3. Custom Key Functions

```go
// Custom key function
customKeyFunc := func(r *http.Request) string {
    userID := r.Header.Get("X-User-ID")
    if userID != "" {
        return "user:" + userID
    }
    return "ip:" + getClientIP(r)
}

// Use custom key function
config := ratelimit.MiddlewareConfig{
    RateLimiter: rateLimiter,
    DefaultRule: ratelimit.RateLimitRule{
        Requests: 100,
        Duration: time.Minute,
    },
    KeyFunc: customKeyFunc,
}
r.Use(ratelimit.RateLimitMiddleware(config))
```

## Configuration

### Environment Variables

```env
# Rate Limiting Configuration
RATE_LIMIT_ENABLED=true
RATE_LIMIT_KEY_PREFIX=ratelimit
RATE_LIMIT_DEFAULT_REQUESTS=100
RATE_LIMIT_DEFAULT_DURATION_MINUTES=1

# Auth Rate Limiting
RATE_LIMIT_AUTH_LOGIN_REQUESTS=5
RATE_LIMIT_AUTH_LOGIN_DURATION_MINUTES=15
RATE_LIMIT_AUTH_REGISTER_REQUESTS=3
RATE_LIMIT_AUTH_REGISTER_DURATION_MINUTES=60

# User Rate Limiting
RATE_LIMIT_USERS_REQUESTS=50
RATE_LIMIT_USERS_DURATION_MINUTES=1

# Upload Rate Limiting
RATE_LIMIT_UPLOAD_REQUESTS=10
RATE_LIMIT_UPLOAD_DURATION_MINUTES=5

# IP Rate Limiting
RATE_LIMIT_IP_GLOBAL_REQUESTS=1000
RATE_LIMIT_IP_GLOBAL_DURATION_MINUTES=60
```

### Route-specific Rules

```go
rules := map[string]ratelimit.RateLimitRule{
    "/api/v1/auth/login": {
        Requests: 5,
        Duration: 15 * time.Minute,
    },
    "/api/v1/auth/register": {
        Requests: 3,
        Duration: time.Hour,
    },
    "/api/v1/users": {
        Requests: 50,
        Duration: time.Minute,
    },
    "/api/v1/upload": {
        Requests: 10,
        Duration: 5 * time.Minute,
    },
}
```

## Response Headers

Rate limiting middleware tự động thêm các headers sau vào response:

- `X-RateLimit-Limit`: Số requests được phép
- `X-RateLimit-Remaining`: Số requests còn lại
- `X-RateLimit-Reset`: Timestamp khi rate limit reset
- `X-RateLimit-Retry-After`: Số giây cần chờ trước khi retry (chỉ khi rate limit exceeded)

## Error Handling

Khi rate limit exceeded, middleware sẽ:

1. Trả về HTTP status `429 Too Many Requests`
2. Thêm `X-RateLimit-Retry-After` header
3. Trả về error message: "Rate limit exceeded"

## Redis Key Structure

Rate limiting sử dụng Redis keys với format:

```
{keyPrefix}:{key}
```

Ví dụ:

- `ratelimit:ip:192.168.1.1`
- `ratelimit:user:123`
- `ratelimit:ip:192.168.1.1:route:/api/v1/auth/login`

## Performance

- Sử dụng Redis pipeline cho atomic operations
- TTL được set tự động để cleanup expired keys
- Efficient memory usage với Redis expiration

## Monitoring

Có thể monitor rate limiting thông qua:

1. Redis keys và TTL
2. Response headers
3. Application logs
4. Redis metrics

## Examples

### Example 1: Basic IP Rate Limiting

```go
// 100 requests per minute per IP
r.Use(ratelimit.RateLimitByIP(rateLimiter, 100, time.Minute))
```

### Example 2: Route-specific Rate Limiting

```go
rules := map[string]ratelimit.RateLimitRule{
    "/api/v1/auth/login": {
        Requests: 5,
        Duration: 15 * time.Minute,
    },
    "/api/v1/users": {
        Requests: 50,
        Duration: time.Minute,
    },
}
r.Use(ratelimit.RateLimitByRoute(rateLimiter, rules))
```

### Example 3: User-based Rate Limiting

```go
// 1000 requests per hour per user
r.Use(ratelimit.RateLimitByUser(rateLimiter, 1000, time.Hour))
```

### Example 4: Using in Routes

```go
// Global rate limiting
r.Use(middleware.RateLimitByIP(redisClient, 1000, time.Hour))

// Auth routes with strict rate limiting
r.Group(func(r chi.Router) {
    r.Use(middleware.RateLimitByIP(redisClient, 5, 15*time.Minute))
    r.Post("/login", authHandler.Login)
    r.Post("/register", authHandler.Register)
})

// User routes with user-based rate limiting
r.Group(func(r chi.Router) {
    r.Use(jwtMiddleware)
    r.Use(middleware.RateLimitByUserOrIP(redisClient, 100, time.Minute))
    r.Get("/users", userHandler.GetUsers)
    r.Post("/users", userHandler.CreateUser)
})
```

### Example 5: Custom Configuration

```go
config := ratelimit.MiddlewareConfig{
    RateLimiter: rateLimiter,
    Rules: map[string]ratelimit.RateLimitRule{
        "/api/v1/auth/login": {
            Requests: 5,
            Duration: 15 * time.Minute,
        },
    },
    DefaultRule: ratelimit.RateLimitRule{
        Requests: 100,
        Duration: time.Minute,
    },
    KeyFunc: ratelimit.KeyByIPAndRoute,
}
r.Use(ratelimit.RateLimitMiddleware(config))
```
