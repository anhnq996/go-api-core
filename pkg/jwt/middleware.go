package jwt

import (
	"context"
	"net/http"
	"strings"

	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/response"
)

// contextKey là kiểu để lưu claims vào context
type contextKey string

const (
	// ClaimsContextKey là key để lưu claims trong context
	ClaimsContextKey contextKey = "jwt_claims"
	// UserIDContextKey là key để lưu user ID trong context
	UserIDContextKey contextKey = "user_id"
)

// Middleware xác thực JWT token
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguageFromContext(r.Context())

		// Lấy token từ Authorization header
		token := extractTokenFromHeader(r)
		if token == "" {
			response.Unauthorized(w, lang, response.CodeTokenMissing)
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

		// Lưu claims vào context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)

		// Tiếp tục với request có context mới
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware xác thực JWT token nhưng không bắt buộc
// Nếu có token hợp lệ thì lưu claims vào context, nếu không có hoặc invalid thì vẫn cho qua
func (m *Manager) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromHeader(r)
		if token != "" {
			claims, err := m.VerifyToken(token)
			if err == nil {
				ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
				ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware kiểm tra role của user
func (m *Manager) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := i18n.GetLanguageFromContext(r.Context())

			claims := GetClaimsFromContext(r.Context())
			if claims == nil {
				response.Unauthorized(w, lang, response.CodeTokenMissing)
				return
			}

			// Kiểm tra role
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.Forbidden(w, lang, response.CodePermissionDenied)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTokenFromHeader lấy token từ Authorization header
func extractTokenFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	// Format: "Bearer <token>"
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetClaimsFromContext lấy claims từ context
func GetClaimsFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserIDFromContext lấy user ID từ context
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// MustGetClaimsFromContext lấy claims từ context, panic nếu không có
func MustGetClaimsFromContext(ctx context.Context) *Claims {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		panic("claims not found in context")
	}
	return claims
}

// MustGetUserIDFromContext lấy user ID từ context, panic nếu không có
func MustGetUserIDFromContext(ctx context.Context) string {
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		panic("user ID not found in context")
	}
	return userID
}
