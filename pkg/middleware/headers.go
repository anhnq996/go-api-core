package middleware

import (
	"net/http"
	"os"
)

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CustomHeaders middleware adds custom headers to responses
func CustomHeaders(headers map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set custom headers
			for key, value := range headers {
				w.Header().Set(key, value)
			}

			// Set additional headers from environment variables
			if apiVersion := getEnv("API_VERSION", ""); apiVersion != "" {
				w.Header().Set("X-API-Version", apiVersion)
			}
			if poweredBy := getEnv("API_POWERED_BY", ""); poweredBy != "" {
				w.Header().Set("X-Powered-By", poweredBy)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// CORSHeaders middleware adds CORS headers manually
func CORSHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers from environment variables
			w.Header().Set("Access-Control-Allow-Origin", getEnv("CORS_ALLOWED_ORIGINS", "*"))
			w.Header().Set("Access-Control-Allow-Methods", getEnv("CORS_ALLOWED_METHODS", "GET, POST, PUT, DELETE, OPTIONS, PATCH"))
			w.Header().Set("Access-Control-Allow-Headers", getEnv("CORS_ALLOWED_HEADERS", "*"))
			w.Header().Set("Access-Control-Expose-Headers", getEnv("CORS_EXPOSED_HEADERS", "Link"))
			w.Header().Set("Access-Control-Max-Age", getEnv("CORS_MAX_AGE", "300"))

			// Set Allow-Credentials if enabled
			if getEnv("CORS_ALLOW_CREDENTIALS", "false") == "true" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight OPTIONS requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}
