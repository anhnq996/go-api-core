# JWT Package

Package JWT cung cấp đầy đủ chức năng để xử lý JSON Web Tokens cho authentication.

## Features

- ✅ Generate access token & refresh token
- ✅ Token verification & parsing
- ✅ JWT middleware cho protected routes
- ✅ Role-based access control (RBAC)
- ✅ Token blacklist (logout functionality)
- ✅ Optional authentication middleware
- ✅ Token refresh mechanism
- ✅ Context helpers
- ✅ Comprehensive error handling

## Installation

```bash
go get github.com/golang-jwt/jwt/v5
```

## Configuration

### Khởi tạo JWT Manager

```go
import (
    "api-core/pkg/jwt"
    "time"
)

// Tạo JWT manager
jwtManager := jwt.NewManager(jwt.Config{
    SecretKey:            "your-secret-key-min-32-characters",
    AccessTokenDuration:  15 * time.Minute,
    RefreshTokenDuration: 7 * 24 * time.Hour,
    Issuer:               "apicore",
})
```

### Environment Variables

```env
# .env
JWT_SECRET_KEY=your-super-secret-key-at-least-32-characters-long
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=168h
```

## Basic Usage

### 1. Generate Tokens

```go
// Generate token pair (access + refresh)
tokens, err := jwtManager.GenerateTokenPair(
    "user123",                    // User ID
    "user@example.com",           // Email
    "admin",                      // Role
    map[string]interface{}{       // Metadata (optional)
        "name": "John Doe",
        "verified": true,
    },
)

// Response
// {
//   "access_token": "eyJhbGc...",
//   "refresh_token": "eyJhbGc...",
//   "expires_at": "2024-01-15T10:45:00Z",
//   "token_type": "Bearer"
// }
```

### 2. Verify Token

```go
claims, err := jwtManager.VerifyToken(tokenString)
if err != nil {
    if err == jwt.ErrExpiredToken {
        // Token hết hạn
    } else if err == jwt.ErrInvalidToken {
        // Token không hợp lệ
    }
    return
}

// Access claims
userID := claims.UserID
email := claims.Email
role := claims.Role
metadata := claims.Metadata
```

### 3. Refresh Token

```go
// Verify refresh token và tạo access token mới
newTokens, err := jwtManager.RefreshAccessToken(
    refreshToken,
    "user@example.com",
    "admin",
    metadata,
)
```

## Middleware Usage

### 1. Protected Routes

```go
import (
    "github.com/go-chi/chi/v5"
)

func setupRoutes(r *chi.Mux, jwtManager *jwt.Manager) {
    // Public routes
    r.Post("/auth/login", LoginHandler)
    r.Post("/auth/register", RegisterHandler)

    // Protected routes
    r.Group(func(r chi.Router) {
        // Apply JWT middleware
        r.Use(jwtManager.Middleware)

        r.Get("/users/me", GetCurrentUser)
        r.Put("/users/me", UpdateCurrentUser)
        r.Get("/orders", GetOrders)
    })
}
```

### 2. Role-Based Access Control

```go
func setupRoutes(r *chi.Mux, jwtManager *jwt.Manager) {
    // Admin only routes
    r.Group(func(r chi.Router) {
        r.Use(jwtManager.Middleware)
        r.Use(jwtManager.RequireRole("admin"))

        r.Get("/admin/users", GetAllUsers)
        r.Delete("/admin/users/{id}", DeleteUser)
    })

    // Admin or moderator routes
    r.Group(func(r chi.Router) {
        r.Use(jwtManager.Middleware)
        r.Use(jwtManager.RequireRole("admin", "moderator"))

        r.Post("/posts/approve", ApprovePost)
    })
}
```

### 3. Optional Authentication

```go
// Route có thể access với hoặc không có token
r.Group(func(r chi.Router) {
    r.Use(jwtManager.OptionalMiddleware)

    // Nếu có token -> user-specific data
    // Nếu không có token -> public data
    r.Get("/posts", GetPosts)
})
```

### 4. Access Claims in Handler

```go
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    // Lấy claims từ context
    claims := jwt.GetClaimsFromContext(r.Context())
    if claims == nil {
        response.Unauthorized(w, lang, response.CodeUnauthorized)
        return
    }

    // Access user info
    userID := claims.UserID
    email := claims.Email
    role := claims.Role

    // Hoặc dùng helper
    userID := jwt.GetUserIDFromContext(r.Context())

    // Get user from database...
    response.Success(w, lang, response.CodeSuccess, user)
}
```

## Token Blacklist (Logout)

### Setup Blacklist

```go
import (
    "api-core/pkg/cache"
    "api-core/pkg/jwt"
)

// Tạo blacklist với cache
cacheClient := cache.NewRedisCache(...)
blacklist := jwt.NewBlacklist(cacheClient)

// Use middleware với blacklist
r.Group(func(r chi.Router) {
    r.Use(jwtManager.MiddlewareWithBlacklist(blacklist))

    r.Get("/protected", ProtectedHandler)
})
```

### Logout (Single Device)

```go
func Logout(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    // Lấy token từ header
    token := r.Header.Get("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")

    // Get token expiry
    expiry, _ := jwtManager.GetTokenExpiry(token)

    // Add to blacklist
    err := blacklist.Add(token, expiry)
    if err != nil {
        response.InternalServerError(w, lang, response.CodeInternalServerError)
        return
    }

    response.Success(w, lang, response.CodeSuccess, nil)
}
```

### Logout All Devices

```go
func LogoutAll(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())
    claims := jwt.GetClaimsFromContext(r.Context())

    // Blacklist tất cả tokens của user
    expiry := time.Now().Add(7 * 24 * time.Hour) // Max token duration
    err := blacklist.AddUserTokens(claims.UserID, expiry)
    if err != nil {
        response.InternalServerError(w, lang, response.CodeInternalServerError)
        return
    }

    response.Success(w, lang, response.CodeSuccess, nil)
}
```

## Complete Authentication Example

### 1. Login Handler

```go
func Login(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    var input struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, lang, response.CodeInvalidInput, nil)
        return
    }

    // Verify credentials
    user, err := userService.GetByEmail(input.Email)
    if err != nil {
        response.Unauthorized(w, lang, response.CodeInvalidCredentials)
        return
    }

    if !utils.CheckPassword(input.Password, user.Password) {
        response.Unauthorized(w, lang, response.CodeInvalidCredentials)
        return
    }

    // Generate tokens
    tokens, err := jwtManager.GenerateTokenPair(
        user.ID,
        user.Email,
        user.Role,
        map[string]interface{}{
            "name": user.Name,
        },
    )
    if err != nil {
        response.InternalServerError(w, lang, response.CodeInternalServerError)
        return
    }

    response.Success(w, lang, response.CodeSuccess, tokens)
}
```

### 2. Refresh Token Handler

```go
func RefreshToken(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    var input struct {
        RefreshToken string `json:"refresh_token"`
    }

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, lang, response.CodeInvalidInput, nil)
        return
    }

    // Verify refresh token
    userID, err := jwtManager.VerifyRefreshToken(input.RefreshToken)
    if err != nil {
        if err == jwt.ErrExpiredToken {
            response.Unauthorized(w, lang, response.CodeTokenExpired)
            return
        }
        response.Unauthorized(w, lang, response.CodeTokenInvalid)
        return
    }

    // Get user info
    user, err := userService.GetByID(userID)
    if err != nil {
        response.NotFound(w, lang, response.CodeUserNotFound)
        return
    }

    // Generate new tokens
    tokens, err := jwtManager.GenerateTokenPair(
        user.ID,
        user.Email,
        user.Role,
        map[string]interface{}{
            "name": user.Name,
        },
    )
    if err != nil {
        response.InternalServerError(w, lang, response.CodeInternalServerError)
        return
    }

    response.Success(w, lang, response.CodeSuccess, tokens)
}
```

### 3. Get Current User

```go
func GetMe(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    // Get user ID from JWT claims
    userID := jwt.GetUserIDFromContext(r.Context())

    // Get user from database
    user, err := userService.GetByID(userID)
    if err != nil {
        response.NotFound(w, lang, response.CodeUserNotFound)
        return
    }

    response.Success(w, lang, response.CodeSuccess, user)
}
```

## Security Best Practices

### 1. Secret Key

```go
// ✅ Good - Strong secret key
SecretKey: "your-super-secret-key-at-least-32-characters-long-with-random-characters-$#@!"

// ❌ Bad - Weak secret key
SecretKey: "secret"
```

### 2. Token Duration

```go
// ✅ Good - Short access token, long refresh token
AccessTokenDuration:  15 * time.Minute
RefreshTokenDuration: 7 * 24 * time.Hour

// ❌ Bad - Too long access token
AccessTokenDuration:  24 * time.Hour
```

### 3. HTTPS Only

```go
// Production - Always use HTTPS
// Set Secure flag for cookies containing tokens
cookie := &http.Cookie{
    Secure:   true,  // Only send over HTTPS
    HttpOnly: true,  // Not accessible via JavaScript
    SameSite: http.SameSiteStrictMode,
}
```

### 4. Token Storage

```go
// ✅ Client-side: Store in httpOnly cookie or memory
// ✅ Never store in localStorage (vulnerable to XSS)

// ❌ localStorage.setItem('token', token)  // BAD!
```

### 5. Refresh Token Rotation

```go
// Khi refresh, invalidate refresh token cũ
func RefreshToken(oldRefreshToken string) (*TokenPair, error) {
    // 1. Verify old refresh token
    // 2. Generate new token pair
    // 3. Blacklist old refresh token
    blacklist.Add(oldRefreshToken, expiry)
    // 4. Return new tokens
}
```

## Error Handling

```go
token, err := jwtManager.VerifyToken(tokenString)
if err != nil {
    switch err {
    case jwt.ErrExpiredToken:
        // Token hết hạn, yêu cầu refresh
        response.Unauthorized(w, lang, response.CodeTokenExpired)
    case jwt.ErrInvalidToken:
        // Token không hợp lệ
        response.Unauthorized(w, lang, response.CodeTokenInvalid)
    case jwt.ErrInvalidSignature:
        // Signature không đúng (có thể bị tamper)
        response.Unauthorized(w, lang, response.CodeTokenInvalid)
    default:
        response.Unauthorized(w, lang, response.CodeUnauthorized)
    }
    return
}
```

## Testing

```go
func TestJWT(t *testing.T) {
    manager := jwt.NewManager(jwt.Config{
        SecretKey:            "test-secret-key-32-characters-min",
        AccessTokenDuration:  15 * time.Minute,
        RefreshTokenDuration: 7 * 24 * time.Hour,
    })

    // Test generate token
    token, err := manager.GenerateToken("user123", "user@test.com", "admin", nil)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    // Test verify token
    claims, err := manager.VerifyToken(token)
    assert.NoError(t, err)
    assert.Equal(t, "user123", claims.UserID)
    assert.Equal(t, "user@test.com", claims.Email)
}
```

## Performance Tips

- Token verification rất nhanh (microseconds)
- Không cần query database cho mỗi request
- Chỉ cần verify signature và expiry
- Blacklist check chỉ cần 1 Redis lookup

## Troubleshooting

### Token always invalid

```go
// Check secret key match
// Generation secret == Verification secret

// Check token format
// Should be: "Bearer eyJhbGc..."
```

### Token expired immediately

```go
// Check system time
// Check AccessTokenDuration config

// Debug: Print expiry time
expiry, _ := manager.GetTokenExpiry(token)
fmt.Println("Token expires at:", expiry)
```

### Middleware not working

```go
// Check middleware order
r.Use(i18n.Middleware)      // First
r.Use(jwtManager.Middleware) // Then JWT

// Check Authorization header format
Authorization: Bearer eyJhbGc...
```

## Resources

- [JWT.io](https://jwt.io) - Debug and decode tokens
- [RFC 7519](https://tools.ietf.org/html/rfc7519) - JWT Specification
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - Go JWT library
