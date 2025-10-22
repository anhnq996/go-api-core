# JWT Authentication Guide

Hướng dẫn sử dụng JWT (JSON Web Tokens) cho authentication trong ApiCore.

## Tổng quan

JWT package cung cấp đầy đủ chức năng:

- ✅ Access token & Refresh token
- ✅ Token verification & parsing
- ✅ Protected routes middleware
- ✅ Role-based access control (RBAC)
- ✅ Token blacklist (logout)
- ✅ Optional authentication
- ✅ Context helpers

## Quick Start

### 1. Cấu hình Environment

```env
# .env
JWT_SECRET_KEY=your-super-secret-key-at-least-32-characters-long-change-this
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=168h
```

⚠️ **Quan trọng**: Secret key phải:

- Ít nhất 32 ký tự
- Random và phức tạp
- Giữ bí mật, không commit vào git
- Khác nhau giữa dev và production

### 2. Khởi tạo trong main.go

```go
import (
    "anhnq/api-core/pkg/jwt"
    "anhnq/api-core/pkg/cache"
    "time"
)

func main() {
    // ... logger, i18n, database setup

    // Initialize JWT manager
    jwtManager := jwt.NewManager(jwt.Config{
        SecretKey:            os.Getenv("JWT_SECRET_KEY"),
        AccessTokenDuration:  15 * time.Minute,
        RefreshTokenDuration: 7 * 24 * time.Hour,
        Issuer:               "apicore",
    })

    // Initialize blacklist (optional, for logout)
    blacklist := jwt.NewBlacklist(cacheClient)

    // Setup routes với JWT middleware
    setupRoutes(r, jwtManager, blacklist)
}
```

### 3. Setup Routes

```go
func setupRoutes(r *chi.Mux, jwtManager *jwt.Manager, blacklist *jwt.Blacklist) {
    // Public routes (không cần authentication)
    r.Post("/auth/login", LoginHandler)
    r.Post("/auth/register", RegisterHandler)
    r.Post("/auth/refresh", RefreshTokenHandler)

    // Protected routes (cần authentication)
    r.Group(func(r chi.Router) {
        r.Use(jwtManager.MiddlewareWithBlacklist(blacklist))

        r.Get("/auth/me", GetMeHandler)
        r.Post("/auth/logout", LogoutHandler)
        r.Get("/users/profile", GetProfileHandler)
        r.Put("/users/profile", UpdateProfileHandler)
    })

    // Admin only routes
    r.Group(func(r chi.Router) {
        r.Use(jwtManager.MiddlewareWithBlacklist(blacklist))
        r.Use(jwtManager.RequireRole("admin"))

        r.Get("/admin/users", GetAllUsersHandler)
        r.Delete("/admin/users/{id}", DeleteUserHandler)
    })
}
```

## API Endpoints

### 1. Login

**POST** `/auth/login`

Request:

```json
{
  "email": "user@example.com",
  "password": "MyPassword123!"
}
```

Response (200):

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Đăng nhập thành công",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-15T11:00:00Z",
    "token_type": "Bearer"
  }
}
```

### 2. Refresh Token

**POST** `/auth/refresh`

Request:

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response (200):

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Làm mới token thành công",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-15T11:15:00Z",
    "token_type": "Bearer"
  }
}
```

### 3. Get Current User

**GET** `/auth/me`

Headers:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Response (200):

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Thành công",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "email": "user@example.com",
    "role": "user",
    "metadata": {
      "name": "John Doe"
    }
  }
}
```

### 4. Logout (Single Device)

**POST** `/auth/logout`

Headers:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Response (200):

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Đăng xuất thành công",
  "data": null
}
```

### 5. Logout All Devices

**POST** `/auth/logout-all`

Headers:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Response (200):

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Đăng xuất tất cả thiết bị thành công",
  "data": null
}
```

## Client Integration

### JavaScript/TypeScript

```typescript
// Login
const login = async (email: string, password: string) => {
  const response = await fetch("/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  const data = await response.json();

  if (data.success) {
    // Store tokens
    localStorage.setItem("access_token", data.data.access_token);
    localStorage.setItem("refresh_token", data.data.refresh_token);
  }

  return data;
};

// Make authenticated request
const getProfile = async () => {
  const token = localStorage.getItem("access_token");

  const response = await fetch("/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return response.json();
};

// Refresh token
const refreshToken = async () => {
  const refresh = localStorage.getItem("refresh_token");

  const response = await fetch("/auth/refresh", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refresh }),
  });

  const data = await response.json();

  if (data.success) {
    localStorage.setItem("access_token", data.data.access_token);
    localStorage.setItem("refresh_token", data.data.refresh_token);
  }

  return data;
};

// Axios interceptor for auto refresh
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (
      error.response?.data?.code === "TOKEN_EXPIRED" &&
      !originalRequest._retry
    ) {
      originalRequest._retry = true;

      const newTokens = await refreshToken();

      if (newTokens.success) {
        originalRequest.headers["Authorization"] =
          "Bearer " + newTokens.data.access_token;
        return axios(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);
```

### Mobile (iOS/Android)

#### iOS (Swift)

```swift
// Store tokens securely in Keychain
func saveToken(token: String) {
    let keychain = KeychainSwift()
    keychain.set(token, forKey: "access_token")
}

// Make authenticated request
func fetchProfile() {
    let token = KeychainSwift().get("access_token")

    var request = URLRequest(url: url)
    request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

    // ... URLSession request
}
```

#### Android (Kotlin)

```kotlin
// Store tokens in EncryptedSharedPreferences
val prefs = EncryptedSharedPreferences.create(...)
prefs.edit().putString("access_token", token).apply()

// Make authenticated request
val token = prefs.getString("access_token", "")
val request = Request.Builder()
    .url(url)
    .addHeader("Authorization", "Bearer $token")
    .build()
```

## Security Best Practices

### 1. Secret Key Security

```bash
# Generate strong secret key
openssl rand -base64 48

# Hoặc trong Go
go run -c 'import "crypto/rand"; import "encoding/base64"; b := make([]byte, 48); rand.Read(b); println(base64.StdEncoding.EncodeToString(b))'
```

### 2. Token Storage

| Storage         | Security   | Recommendation         |
| --------------- | ---------- | ---------------------- |
| httpOnly Cookie | ⭐⭐⭐⭐⭐ | Best for web apps      |
| Memory/State    | ⭐⭐⭐⭐   | Good for SPAs          |
| localStorage    | ⭐⭐       | Avoid (XSS vulnerable) |
| sessionStorage  | ⭐⭐       | Avoid (XSS vulnerable) |

### 3. Token Expiry

```go
// ✅ Good - Short access token
AccessTokenDuration: 15 * time.Minute

// ✅ Good - Long refresh token
RefreshTokenDuration: 7 * 24 * time.Hour

// ❌ Bad - Too long access token
AccessTokenDuration: 24 * time.Hour
```

### 4. HTTPS Only

```go
// Production - Always use HTTPS
if isProduction {
    // Force HTTPS middleware
    r.Use(middleware.ForceHTTPS)
}
```

### 5. Rate Limiting

```go
// Limit login attempts
r.With(rateLimiter.Limit("5-M")).Post("/auth/login", LoginHandler)
```

## Error Handling

| Error Code        | HTTP Status | Message (EN)      | Message (VI)               |
| ----------------- | ----------- | ----------------- | -------------------------- |
| TOKEN_MISSING     | 401         | Token is required | Thiếu token xác thực       |
| TOKEN_INVALID     | 401         | Invalid token     | Token không hợp lệ         |
| TOKEN_EXPIRED     | 401         | Token has expired | Phiên đăng nhập đã hết hạn |
| PERMISSION_DENIED | 403         | Permission denied | Không có quyền truy cập    |

## Testing

### Test JWT with curl

```bash
# 1. Login
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Response: Get access_token

# 2. Access protected route
curl http://localhost:3000/auth/me \
  -H "Authorization: Bearer <access_token>"

# 3. Refresh token
curl -X POST http://localhost:3000/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'

# 4. Logout
curl -X POST http://localhost:3000/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

### Test với Postman

1. **Login**:

   - POST `/auth/login`
   - Body: `{"email":"...", "password":"..."}`
   - Save `access_token` to environment variable

2. **Protected Request**:
   - GET `/auth/me`
   - Headers: `Authorization: Bearer {{access_token}}`

## Troubleshooting

### Problem: Token always invalid

**Cause**: Secret key không match
**Solution**: Kiểm tra JWT_SECRET_KEY trong .env

### Problem: Token expired immediately

**Cause**: System time không sync
**Solution**: Sync system time hoặc tăng token duration

### Problem: Middleware not working

**Cause**: Wrong middleware order
**Solution**:

```go
r.Use(i18n.Middleware)       // First
r.Use(jwtManager.Middleware) // Then JWT
```

### Problem: Cannot access user info in handler

**Cause**: Chưa dùng JWT middleware
**Solution**: Check route có wrap trong `r.Use(jwtManager.Middleware)`

## Advanced Usage

### Custom Claims

```go
type CustomClaims struct {
    UserID      string `json:"user_id"`
    Email       string `json:"email"`
    Role        string `json:"role"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}
```

### Multiple Roles Check

```go
func hasAnyRole(claims *jwt.Claims, roles ...string) bool {
    for _, role := range roles {
        if claims.Role == role {
            return true
        }
    }
    return false
}
```

### Token Refresh Strategy

```go
// Client-side: Check expiry before request
if (tokenExpiry - now < 5 minutes) {
    await refreshToken();
}

// Or: Auto refresh on 401 response
```

## Performance

- Token generation: ~0.1ms
- Token verification: ~0.05ms
- Blacklist check: ~1ms (Redis lookup)
- Total overhead: ~1-2ms per request

## Resources

- [JWT.io](https://jwt.io) - Debug tokens
- [RFC 7519](https://tools.ietf.org/html/rfc7519) - JWT Spec
- [pkg/jwt/README.md](../pkg/jwt/README.md) - Package documentation
