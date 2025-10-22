package i18n

import (
	"context"
	"net/http"
)

// contextKey là kiểu để lưu language vào context
type contextKey string

const (
	// LanguageContextKey là key để lưu language trong context
	LanguageContextKey contextKey = "language"
)

// Middleware tự động parse và lưu language vào context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := detectLanguage(r)

		// Lưu language vào context
		ctx := context.WithValue(r.Context(), LanguageContextKey, lang)

		// Tiếp tục với request có context mới
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// detectLanguage phát hiện ngôn ngữ từ request
func detectLanguage(r *http.Request) string {
	// 1. Kiểm tra query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if HasLanguage(lang) {
			return lang
		}
	}

	// 2. Kiểm tra Accept-Language header
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		lang := ParseAcceptLanguage(acceptLang)
		if lang != "" && HasLanguage(lang) {
			return lang
		}
	}

	// 3. Default
	return "en"
}

// GetLanguageFromContext lấy language từ context
func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageContextKey).(string); ok {
		return lang
	}
	return "en"
}
