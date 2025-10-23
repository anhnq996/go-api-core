package exception

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/response"

	"github.com/go-chi/chi/v5/middleware"
)

// RecoveryMiddleware recovers from panics and converts them to proper HTTP responses
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for logging
				requestID := middleware.GetReqID(r.Context())

				// Log the panic
				stack := debug.Stack()
				fmt.Printf("PANIC [%s]: %v\n%s", requestID, err, stack)

				// Log the exception details
				_ = NewWithCode("Internal server error", "INTERNAL_ERROR").
					WithContext("request_id", requestID).
					WithContext("panic", err).
					WithContext("stack", string(stack))

				// Send error response
				lang := i18n.GetLanguageFromContext(r.Context())
				response.InternalServerError(w, lang, response.CodeInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ExceptionHandler middleware handles exceptions and converts them to proper HTTP responses
func ExceptionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware can be used to catch and handle exceptions
		// that are thrown during request processing
		next.ServeHTTP(w, r)
	})
}

// HandleException converts an exception to HTTP response
func HandleException(w http.ResponseWriter, r *http.Request, err error) {
	lang := i18n.GetLanguageFromContext(r.Context())

	if ex, ok := err.(*Exception); ok {
		// Handle custom exception
		switch ex.Code {
		case "NOT_FOUND":
			response.NotFound(w, lang, response.CodeNotFound)
		case "UNAUTHORIZED":
			response.Unauthorized(w, lang, response.CodeUnauthorized)
		case "FORBIDDEN":
			response.Forbidden(w, lang, response.CodeForbidden)
		case "BAD_REQUEST":
			response.BadRequest(w, lang, response.CodeBadRequest, nil)
		case "VALIDATION_ERROR":
			response.BadRequest(w, lang, response.CodeValidationFailed, nil)
		case "CONFLICT":
			response.Conflict(w, lang, response.CodeConflict)
		case "TIMEOUT":
			response.InternalServerError(w, lang, response.CodeInternalServerError)
		default:
			response.InternalServerError(w, lang, response.CodeInternalServerError)
		}
	} else {
		// Handle regular error
		response.InternalServerError(w, lang, response.CodeInternalServerError)
	}
}

// PanicHandler handles panics in handlers
func PanicHandler(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for logging
				requestID := middleware.GetReqID(r.Context())

				// Log the panic
				stack := debug.Stack()
				fmt.Printf("PANIC in handler [%s]: %v\n%s", requestID, err, stack)

				// Send error response
				lang := i18n.GetLanguageFromContext(r.Context())
				response.InternalServerError(w, lang, response.CodeInternalServerError)
			}
		}()

		handler(w, r)
	}
}

// SafeHandler wraps a handler with panic recovery
func SafeHandler(handler http.HandlerFunc) http.HandlerFunc {
	return PanicHandler(handler)
}
