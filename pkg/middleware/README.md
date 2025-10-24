# Middleware Package

Package này cung cấp các middleware hữu ích cho ứng dụng ApiCore.

## Các Middleware có sẵn:

### 1. CustomHeaders

Thêm custom headers vào response:

```go
import "anhnq/api-core/pkg/middleware"

// Thêm custom headers
r.Use(middleware.CustomHeaders(map[string]string{
    "X-API-Version": "1.0",
    "X-Powered-By":  "ApiCore",
    "X-Custom-Header": "Custom Value",
}))
```

### 2. CORSHeaders

Thêm CORS headers vào response:

```go
import "anhnq/api-core/pkg/middleware"

// Thêm CORS headers
r.Use(middleware.CORSHeaders())
```

### 3. SecurityHeaders

Thêm security headers vào response:

```go
import "anhnq/api-core/pkg/middleware"

// Thêm security headers
r.Use(middleware.SecurityHeaders())
```

## Cách sử dụng trong Controller:

### 1. Set headers trực tiếp trong controller:

```go
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
    // Set custom headers
    w.Header().Set("X-Custom-Header", "Custom Value")
    w.Header().Set("X-API-Version", "1.0")

    // Set CORS headers manually
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

    // Response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}
```

### 2. Set headers cho specific endpoint:

```go
// Route với custom headers
r.Route("/api/v1", func(r chi.Router) {
    r.Use(middleware.CustomHeaders(map[string]string{
        "X-API-Version": "1.0",
        "X-API-Name": "ApiCore",
    }))

    r.Get("/users", handler.GetUsers)
    r.Post("/users", handler.CreateUser)
})
```

### 3. Set headers cho specific method:

```go
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
    // Set headers based on request method
    if r.Method == "GET" {
        w.Header().Set("X-Response-Type", "List")
    } else if r.Method == "POST" {
        w.Header().Set("X-Response-Type", "Created")
    }

    // Response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}
```

## Headers được thêm tự động:

### CORS Headers (từ environment variables):

- `Access-Control-Allow-Origin: *` (từ `CORS_ALLOWED_ORIGINS`)
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH` (từ `CORS_ALLOWED_METHODS`)
- `Access-Control-Allow-Headers: *` (từ `CORS_ALLOWED_HEADERS`)
- `Access-Control-Expose-Headers: Link` (từ `CORS_EXPOSED_HEADERS`)
- `Access-Control-Max-Age: 300` (từ `CORS_MAX_AGE`)
- `Access-Control-Allow-Credentials: true` (nếu `CORS_ALLOW_CREDENTIALS=true`)

### Security Headers:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

### Custom Headers (từ environment variables):

- `X-API-Version: 1.0` (từ `API_VERSION`)
- `X-Powered-By: ApiCore` (từ `API_POWERED_BY`)

## Environment Variables:

### CORS Configuration:

```env
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS,PATCH
CORS_ALLOWED_HEADERS=*
CORS_EXPOSED_HEADERS=Link
CORS_ALLOW_CREDENTIALS=false
CORS_MAX_AGE=300
```

### API Headers Configuration:

```env
API_VERSION=1.0
API_POWERED_BY=ApiCore
```

## Lưu ý:

1. **Thứ tự middleware**: Middleware được áp dụng theo thứ tự được đăng ký
2. **Headers override**: Headers được set sau sẽ override headers được set trước
3. **CORS**: CORS headers phải được set trước khi response được gửi
4. **Security**: Security headers nên được set cho tất cả endpoints
