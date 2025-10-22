package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP lấy IP address của client
func GetClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For có thể chứa nhiều IPs, lấy cái đầu tiên
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// Try X-Real-IP header
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// GetUserAgent lấy User-Agent string
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// IsAjax kiểm tra có phải AJAX request không
func IsAjax(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// GetQueryParam lấy query parameter với default value
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryParamInt lấy query parameter dạng int
func GetQueryParamInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue := ToInt(value)
	if intValue == 0 {
		return defaultValue
	}

	return intValue
}

// SetCookie set cookie
func SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false, // Set true nếu dùng HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetCookie lấy cookie value
func GetCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// DeleteCookie xóa cookie
func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// GetReferer lấy referer URL
func GetReferer(r *http.Request) string {
	return r.Header.Get("Referer")
}

// IsJSON kiểm tra request có Content-Type JSON không
func IsJSONRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json")
}

// GetAcceptLanguage lấy Accept-Language header
func GetAcceptLanguage(r *http.Request) string {
	return r.Header.Get("Accept-Language")
}

// GetBearerToken lấy Bearer token từ Authorization header
func GetBearerToken(r *http.Request) string {
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

// SetJSONContentType set Content-Type header
func SetJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// SetNoCacheHeaders set no-cache headers
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
