package jwt

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"anhnq/api-core/pkg/cache"
	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/response"
)

// Blacklist quản lý danh sách tokens bị blacklist (logout)
type Blacklist struct {
	cache  cache.Cache
	prefix string
}

// NewBlacklist tạo blacklist mới
func NewBlacklist(c cache.Cache) *Blacklist {
	return &Blacklist{
		cache:  c,
		prefix: "jwt:blacklist:",
	}
}

// Add thêm token vào blacklist
func (b *Blacklist) Add(token string, expiry time.Time) error {
	key := b.prefix + token
	ttl := time.Until(expiry)

	if ttl <= 0 {
		// Token đã hết hạn, không cần blacklist
		return nil
	}

	return b.cache.Set(context.Background(), key, "1", ttl)
}

// IsBlacklisted kiểm tra token có trong blacklist không
func (b *Blacklist) IsBlacklisted(token string) bool {
	key := b.prefix + token
	_, err := b.cache.Get(context.Background(), key)
	return err == nil // Nếu tìm thấy trong cache = blacklisted
}

// Remove xóa token khỏi blacklist (ít dùng)
func (b *Blacklist) Remove(token string) error {
	key := b.prefix + token
	return b.cache.Del(context.Background(), key)
}

// AddUserTokens blacklist tất cả tokens của user (logout all devices)
func (b *Blacklist) AddUserTokens(userID string, expiry time.Time) error {
	key := fmt.Sprintf("jwt:user:blacklist:%s", userID)
	ttl := time.Until(expiry)

	if ttl <= 0 {
		return nil
	}

	return b.cache.Set(context.Background(), key, "1", ttl)
}

// IsUserBlacklisted kiểm tra user có bị blacklist không
func (b *Blacklist) IsUserBlacklisted(userID string) bool {
	key := fmt.Sprintf("jwt:user:blacklist:%s", userID)
	_, err := b.cache.Get(context.Background(), key)
	return err == nil
}

// MiddlewareWithBlacklist middleware kết hợp JWT verification và blacklist check
func (m *Manager) MiddlewareWithBlacklist(blacklist *Blacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := i18n.GetLanguageFromContext(r.Context())

			// Lấy token từ header
			token := ExtractTokenFromHeader(r)
			if token == "" {
				response.Unauthorized(w, lang, response.CodeTokenMissing)
				return
			}

			// Kiểm tra blacklist
			if blacklist.IsBlacklisted(token) {
				response.Unauthorized(w, lang, response.CodeTokenInvalid)
				return
			}

			// Verify token
			claims, err := m.VerifyToken(token)
			if err != nil {
				if err == ErrExpiredToken {
					response.Unauthorized(w, lang, response.CodeTokenExpired)
					return
				}
				response.Unauthorized(w, lang, response.CodeTokenInvalid)
				return
			}

			// Kiểm tra user có bị blacklist không
			if blacklist.IsUserBlacklisted(claims.UserID) {
				response.Unauthorized(w, lang, response.CodeTokenInvalid)
				return
			}

			// Lưu claims vào context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
