package exception

import (
	"api-core/pkg/i18n"
	"api-core/pkg/logger"
	"api-core/pkg/response"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
)

// RecoveryMiddleware recovers from panics and converts them to proper HTTP responses
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				// Get request ID for logging
				requestID := middleware.GetReqID(r.Context())
				jobLogger := logger.GetJobLogger("exception")

				// Log the panic
				stack := debug.Stack()

				logger.Errorf("PANIC [%s]: %v\n%s", requestID, panicErr, stack)

				// Use the already initialized logger
				jobLogger.Error().Msgf("PANIC [%s]: %v\n%s", requestID, panicErr, stack)

				// Determine response code
				var responseCode string
				if ex, ok := panicErr.(Exception); ok {
					// Use exception code if available
					responseCode = ex.Code
				} else {
					// Use default internal server error code
					responseCode = response.CodeInternalServerError
				}

				// Log the exception details
				_ = NewWithCode("Internal server error", responseCode).
					WithContext("request_id", requestID).
					WithContext("panic", panicErr).
					WithContext("stack", string(stack))

				// Send error response
				lang := i18n.GetLanguageFromContext(r.Context())
				response.InternalServerError(w, lang, responseCode)
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
		// Determine response code - ưu tiên ex.Code
		responseCode := response.CodeInternalServerError
		if ex.Code != "" {
			responseCode = ex.Code
		}

		// Handle custom exception
		switch ex.Code {
		case "NOT_FOUND":
			response.NotFound(w, lang, responseCode)
		case "UNAUTHORIZED":
			response.Unauthorized(w, lang, responseCode)
		case "FORBIDDEN":
			response.Forbidden(w, lang, responseCode)
		case "BAD_REQUEST":
			response.BadRequest(w, lang, responseCode, nil)
		case "VALIDATION_ERROR":
			response.BadRequest(w, lang, responseCode, nil)
		case "CONFLICT":
			response.Conflict(w, lang, responseCode)
		case "TIMEOUT":
			response.InternalServerError(w, lang, responseCode)
		default:
			response.InternalServerError(w, lang, responseCode)
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
			if panicErr := recover(); panicErr != nil {
				// Get request ID for logging
				requestID := middleware.GetReqID(r.Context())

				// Log the panic
				stack := debug.Stack()
				fmt.Printf("PANIC in handler [%s]: %v\n%s", requestID, panicErr, stack)

				// Determine response code
				var responseCode string
				if ex, ok := panicErr.(*Exception); ok && ex.Code != "" {
					fmt.Printf("Exception code: %s\n", ex.Code)
					// Use exception code if available
					responseCode = ex.Code
				} else {
					// Use default internal server error code
					responseCode = response.CodeInternalServerError
				}

				// Send error response
				lang := i18n.GetLanguageFromContext(r.Context())
				response.InternalServerError(w, lang, responseCode)
			}
		}()

		handler(w, r)
	}
}

// SafeHandler wraps a handler with panic recovery
func SafeHandler(handler http.HandlerFunc) http.HandlerFunc {
	return PanicHandler(handler)
}
