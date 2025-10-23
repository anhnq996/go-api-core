package exception

import (
	"fmt"
	"runtime"
)

// Exception represents a custom exception with stack trace
type Exception struct {
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	StackTrace []string               `json:"stack_trace,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Inner      error                  `json:"-"`
}

// Error implements the error interface
func (e *Exception) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Inner)
	}
	return e.Message
}

// Unwrap returns the inner error
func (e *Exception) Unwrap() error {
	return e.Inner
}

// New creates a new exception
func New(message string) *Exception {
	return &Exception{
		Message:    message,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// NewWithCode creates a new exception with code
func NewWithCode(message, code string) *Exception {
	return &Exception{
		Message:    message,
		Code:       code,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// Wrap wraps an existing error
func Wrap(err error, message string) *Exception {
	if err == nil {
		return nil
	}

	return &Exception{
		Message:    message,
		Inner:      err,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// WrapWithCode wraps an existing error with code
func WrapWithCode(err error, message, code string) *Exception {
	if err == nil {
		return nil
	}

	return &Exception{
		Message:    message,
		Code:       code,
		Inner:      err,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// WithContext adds context to the exception
func (e *Exception) WithContext(key string, value interface{}) *Exception {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithContexts adds multiple contexts to the exception
func (e *Exception) WithContexts(contexts map[string]interface{}) *Exception {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range contexts {
		e.Context[k] = v
	}
	return e
}

// getStackTrace captures the current stack trace
func getStackTrace() []string {
	var stack []string
	for i := 1; i < 10; i++ { // Skip first few frames
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}
	return stack
}

// Common exception types
var (
	ErrNotFound     = NewWithCode("Resource not found", "NOT_FOUND")
	ErrUnauthorized = NewWithCode("Unauthorized access", "UNAUTHORIZED")
	ErrForbidden    = NewWithCode("Access forbidden", "FORBIDDEN")
	ErrBadRequest   = NewWithCode("Bad request", "BAD_REQUEST")
	ErrInternal     = NewWithCode("Internal server error", "INTERNAL_ERROR")
	ErrValidation   = NewWithCode("Validation failed", "VALIDATION_ERROR")
	ErrConflict     = NewWithCode("Resource conflict", "CONFLICT")
	ErrTimeout      = NewWithCode("Request timeout", "TIMEOUT")
)

// IsException checks if error is an Exception
func IsException(err error) bool {
	_, ok := err.(*Exception)
	return ok
}

// GetExceptionCode returns the exception code if available
func GetExceptionCode(err error) string {
	if ex, ok := err.(*Exception); ok {
		return ex.Code
	}
	return ""
}

// GetExceptionMessage returns the exception message
func GetExceptionMessage(err error) string {
	if ex, ok := err.(*Exception); ok {
		return ex.Message
	}
	return err.Error()
}

// GetExceptionContext returns the exception context
func GetExceptionContext(err error) map[string]interface{} {
	if ex, ok := err.(*Exception); ok {
		return ex.Context
	}
	return nil
}

// GetExceptionStackTrace returns the exception stack trace
func GetExceptionStackTrace(err error) []string {
	if ex, ok := err.(*Exception); ok {
		return ex.StackTrace
	}
	return nil
}
