package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter để capture response
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Write to buffer for logging
	rw.body.Write(b)
	// Write to actual response
	return rw.ResponseWriter.Write(b)
}

// Middleware tạo HTTP logging middleware với đầy đủ request/response
func Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID
			reqID := middleware.GetReqID(r.Context())

			// Read request body
			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Wrap response writer
			ww := newResponseWriter(w)

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			var logEvent *zerolog.Event
			statusCode := ww.statusCode

			if statusCode >= 500 {
				logEvent = RequestLogger.Error()
			} else if statusCode >= 400 {
				logEvent = RequestLogger.Warn()
			} else {
				logEvent = RequestLogger.Info()
			}

			// Add basic request info
			logEvent = logEvent.
				Str("request_id", reqID).
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Str("content_type", r.Header.Get("Content-Type")).
				Int("status", statusCode).
				Dur("duration", duration).
				Int64("duration_ms", duration.Milliseconds())

			// Add request headers (selected)
			logEvent = logEvent.
				Str("accept", r.Header.Get("Accept")).
				Str("referer", r.Header.Get("Referer"))

			// Add request body if present and not too large
			if len(requestBody) > 0 && len(requestBody) < 10000 {
				logEvent = logEvent.
					Str("request_body", string(requestBody)).
					Int("request_size", len(requestBody))
			} else if len(requestBody) > 0 {
				logEvent = logEvent.Int("request_size", len(requestBody))
			}

			// Add response body if present and not too large and not binary
			responseContentType := w.Header().Get("Content-Type")
			if ww.body.Len() > 0 && !isBinaryContent(responseContentType) {
				if ww.body.Len() < 10000 {
					logEvent = logEvent.
						Str("response_body", ww.body.String()).
						Int("response_size", ww.body.Len())
				} else {
					logEvent = logEvent.Int("response_size", ww.body.Len())
				}
			} else if ww.body.Len() > 0 {
				logEvent = logEvent.Int("response_size", ww.body.Len())
			}

			// Add response content type
			logEvent = logEvent.Str("response_content_type", w.Header().Get("Content-Type"))

			// Log message based on status code
			msg := "Request completed"
			if statusCode >= 500 {
				logEvent.Msg("Server error - " + msg)
			} else if statusCode >= 400 {
				logEvent.Msg("Client error - " + msg)
			} else {
				logEvent.Msg(msg)
			}
		})
	}
}

// SimpleMiddleware tạo middleware đơn giản hơn (không log body)
func SimpleMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID
			reqID := middleware.GetReqID(r.Context())

			// Wrap response writer
			ww := newResponseWriter(w)

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request info
			RequestLogger.Info().
				Str("request_id", reqID).
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", ww.statusCode).
				Dur("duration", duration).
				Int64("duration_ms", duration.Milliseconds()).
				Msg("Request completed")
		})
	}
}

// RequestLog log thông tin cơ bản của request (dùng Logger thông thường)
func RequestLog(r *http.Request, msg string) {
	reqID := middleware.GetReqID(r.Context())
	Logger.Info().
		Str("request_id", reqID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg(msg)
}

// RequestLogWithFields log request với custom fields (dùng Logger thông thường)
func RequestLogWithFields(r *http.Request, msg string, fields map[string]interface{}) {
	reqID := middleware.GetReqID(r.Context())
	event := Logger.Info().
		Str("request_id", reqID).
		Str("method", r.Method).
		Str("path", r.URL.Path)

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg(msg)
}

// ErrorLog log error trong request (dùng Logger thông thường)
func ErrorLog(r *http.Request, err error, msg string) {
	reqID := middleware.GetReqID(r.Context())
	Logger.Error().
		Err(err).
		Str("request_id", reqID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg(msg)
}

// MiddlewareConfig cấu hình cho logging middleware
type MiddlewareConfig struct {
	LogRequestBody  bool // Log request body
	LogResponseBody bool // Log response body
	LogHeaders      bool // Log request headers
	MaxBodySize     int  // Max body size to log (bytes)
}

// DefaultMiddlewareConfig cấu hình mặc định
var DefaultMiddlewareConfig = MiddlewareConfig{
	LogRequestBody:  true,
	LogResponseBody: true,
	LogHeaders:      true,
	MaxBodySize:     10000, // 10KB
}

// MiddlewareWithConfig tạo middleware với custom config
func MiddlewareWithConfig(config MiddlewareConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID
			reqID := middleware.GetReqID(r.Context())

			// Read request body if configured
			var requestBody []byte
			if config.LogRequestBody && r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Wrap response writer
			ww := newResponseWriter(w)

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			var logEvent *zerolog.Event
			statusCode := ww.statusCode

			if statusCode >= 500 {
				logEvent = RequestLogger.Error()
			} else if statusCode >= 400 {
				logEvent = RequestLogger.Warn()
			} else {
				logEvent = RequestLogger.Info()
			}

			// Add basic request info
			logEvent = logEvent.
				Str("request_id", reqID).
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", statusCode).
				Dur("duration", duration).
				Int64("duration_ms", duration.Milliseconds())

			// Add headers if configured
			if config.LogHeaders {
				logEvent = logEvent.
					Str("user_agent", r.UserAgent()).
					Str("content_type", r.Header.Get("Content-Type")).
					Str("accept", r.Header.Get("Accept")).
					Str("referer", r.Header.Get("Referer"))
			}

			// Add request body if configured
			if config.LogRequestBody && len(requestBody) > 0 {
				if len(requestBody) < config.MaxBodySize {
					logEvent = logEvent.
						Str("request_body", string(requestBody)).
						Int("request_size", len(requestBody))
				} else {
					logEvent = logEvent.
						Int("request_size", len(requestBody)).
						Str("request_body", "Body too large to log")
				}
			}

			// Add response body if configured and not binary
			responseContentType := w.Header().Get("Content-Type")
			if config.LogResponseBody && ww.body.Len() > 0 && !isBinaryContent(responseContentType) {
				if ww.body.Len() < config.MaxBodySize {
					logEvent = logEvent.
						Str("response_body", ww.body.String()).
						Int("response_size", ww.body.Len())
				} else {
					logEvent = logEvent.
						Int("response_size", ww.body.Len()).
						Str("response_body", "Body too large to log")
				}
			} else if config.LogResponseBody && ww.body.Len() > 0 {
				logEvent = logEvent.Int("response_size", ww.body.Len())
			}

			// Add response content type
			logEvent = logEvent.Str("response_content_type", w.Header().Get("Content-Type"))

			// Log message based on status code
			msg := "Request completed"
			if statusCode >= 500 {
				logEvent.Msg("Server error - " + msg)
			} else if statusCode >= 400 {
				logEvent.Msg("Client error - " + msg)
			} else {
				logEvent.Msg(msg)
			}
		})
	}
}

// isBinaryContent checks if content type is binary
func isBinaryContent(contentType string) bool {
	binaryTypes := []string{
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // Excel
		"application/vnd.ms-excel", // Excel
		"application/pdf",          // PDF
		"image/",                   // Images
		"video/",                   // Videos
		"audio/",                   // Audio
		"application/zip",          // ZIP
		"application/octet-stream", // Binary
		"application/x-",           // Binary applications
		"text/csv",                 // CSV (can be large)
	}

	for _, binaryType := range binaryTypes {
		if len(binaryType) > 0 && binaryType[len(binaryType)-1] == '/' {
			// Check prefix for types like "image/", "video/"
			if len(contentType) >= len(binaryType) && contentType[:len(binaryType)] == binaryType {
				return true
			}
		} else {
			// Exact match for specific types
			if contentType == binaryType {
				return true
			}
		}
	}

	return false
}
