# Exception Package

Package exception provides comprehensive error handling with stack traces, context, and HTTP middleware integration.

## Features

- **Custom Exception Types**: Create exceptions with messages, codes, and context
- **Stack Trace Capture**: Automatic stack trace capture for debugging
- **Context Support**: Add contextual information to exceptions
- **HTTP Middleware**: Recovery middleware and exception handlers
- **Predefined Exceptions**: Common exception types ready to use
- **Error Wrapping**: Wrap existing errors with additional context

## Usage

### Basic Exception Creation

```go
import "api-core/pkg/exception"

// Create a simple exception
ex := exception.New("Something went wrong")

// Create exception with code
ex := exception.NewWithCode("User not found", "USER_NOT_FOUND")

// Wrap an existing error
err := errors.New("database connection failed")
wrappedErr := exception.Wrap(err, "Failed to process request")
```

### Adding Context

```go
ex := exception.New("Validation failed").
    WithContext("field", "email").
    WithContext("value", "invalid-email").
    WithContext("user_id", "123")

// Add multiple contexts at once
contexts := map[string]interface{}{
    "field": "email",
    "value": "invalid-email",
    "user_id": "123",
}
ex := exception.New("Validation failed").WithContexts(contexts)
```

### Predefined Exceptions

```go
// Use predefined exceptions
panic(exception.ErrNotFound.WithContext("resource", "user"))
panic(exception.ErrUnauthorized.WithContext("action", "delete"))
panic(exception.ErrBadRequest.WithContext("field", "email"))
```

### HTTP Middleware

```go
import (
    "github.com/go-chi/chi/v5"
    "api-core/pkg/exception"
)

func main() {
    r := chi.NewRouter()

    // Add recovery middleware
    r.Use(exception.RecoveryMiddleware)

    // Add exception handler middleware
    r.Use(exception.ExceptionHandler)

    // Your routes...
    r.Get("/users", exception.SafeHandler(handleUsers))
}
```

### Safe Handler

```go
func handleUsers(w http.ResponseWriter, r *http.Request) {
    // This handler is automatically protected by panic recovery
    // Any panic will be caught and converted to proper HTTP response

    if r.URL.Query().Get("id") == "" {
        panic(exception.ErrBadRequest.WithContext("missing", "id parameter"))
    }

    // Your handler logic here...
}
```

### Exception Information

```go
err := someFunction()

// Check if error is an exception
if exception.IsException(err) {
    code := exception.GetExceptionCode(err)
    message := exception.GetExceptionMessage(err)
    context := exception.GetExceptionContext(err)
    stackTrace := exception.GetExceptionStackTrace(err)

    fmt.Printf("Code: %s\n", code)
    fmt.Printf("Message: %s\n", message)
    fmt.Printf("Context: %+v\n", context)
    fmt.Printf("Stack Trace: %+v\n", stackTrace)
}
```

## Exception Types

| Exception         | Code               | Description           |
| ----------------- | ------------------ | --------------------- |
| `ErrNotFound`     | `NOT_FOUND`        | Resource not found    |
| `ErrUnauthorized` | `UNAUTHORIZED`     | Unauthorized access   |
| `ErrForbidden`    | `FORBIDDEN`        | Access forbidden      |
| `ErrBadRequest`   | `BAD_REQUEST`      | Bad request           |
| `ErrInternal`     | `INTERNAL_ERROR`   | Internal server error |
| `ErrValidation`   | `VALIDATION_ERROR` | Validation failed     |
| `ErrConflict`     | `CONFLICT`         | Resource conflict     |
| `ErrTimeout`      | `TIMEOUT`          | Request timeout       |

## HTTP Status Mapping

The middleware automatically maps exception codes to appropriate HTTP status codes:

- `NOT_FOUND` → 404 Not Found
- `UNAUTHORIZED` → 401 Unauthorized
- `FORBIDDEN` → 403 Forbidden
- `BAD_REQUEST` → 400 Bad Request
- `VALIDATION_ERROR` → 400 Bad Request
- `CONFLICT` → 409 Conflict
- `TIMEOUT` → 408 Request Timeout
- `INTERNAL_ERROR` → 500 Internal Server Error

## Best Practices

1. **Use Specific Exception Types**: Use predefined exceptions when possible
2. **Add Context**: Always add relevant context to exceptions
3. **Wrap Errors**: Wrap lower-level errors with higher-level context
4. **Use Middleware**: Always use recovery middleware in production
5. **Log Exceptions**: Log exceptions with full context for debugging
6. **Don't Panic in Libraries**: Only panic in handlers, not in service/repository layers

## Example Integration

```go
// In your service layer
func (s *UserService) GetUser(id string) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, exception.WrapWithCode(err, "Failed to get user", "USER_FETCH_ERROR").
            WithContext("user_id", id)
    }

    if user == nil {
        return nil, exception.ErrNotFound.WithContext("resource", "user").
            WithContext("user_id", id)
    }

    return user, nil
}

// In your handler
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    user, err := h.service.GetUser(id)
    if err != nil {
        exception.HandleException(w, r, err)
        return
    }

    response.Success(w, lang, response.CodeSuccess, user)
}
```
