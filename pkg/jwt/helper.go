package jwt

import (
	"net/http"
	"strings"
)

// ExtractTokenFromHeader lấy token từ Authorization header
func ExtractTokenFromHeader(r *http.Request) string {
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
