# Logger Package

Package logger cung cấp structured logging với zerolog và middleware để log HTTP requests/responses.

## Tính Năng

- ✅ Structured logging với zerolog
- ✅ Multiple output (console, file, both)
- ✅ Pretty print cho console (có màu sắc)
- ✅ Log levels: debug, info, warn, error, fatal
- ✅ Middleware để log requests/responses
- ✅ Request ID tracking
- ✅ Log duration của requests
- ✅ Log request body và response body (có giới hạn kích thước)
- ✅ Tự động tạo log file và directory
- ✅ Helper functions tiện lợi

## Cài Đặt

Package đã được tích hợp sẵn trong project.

## Cấu Hình

### Khởi tạo Logger

```go
import "api-core/pkg/logger"

func main() {
    // Khởi tạo logger
    if err := logger.Init(logger.Config{
        Level:        "debug",        // debug, info, warn, error
        Output:       "both",         // console, file, both
        FilePath:     "storages/logs/app.log",
        RequestLogPath: "storages/logs/request.log", // request log path
        EnableCaller: false,          // hiển thị file:line
        PrettyPrint:  true,           // format đẹp cho console
        DailyRotation: true,          // bật daily rotation
    }); err != nil {
        panic(err)
    }
}
```

### Config Options

| Field            | Type   | Description           | Values                             |
| ---------------- | ------ | --------------------- | ---------------------------------- |
| `Level`          | string | Log level             | `debug`, `info`, `warn`, `error`   |
| `Output`         | string | Output destination    | `console`, `file`, `both`          |
| `FilePath`       | string | Log file path         | Ví dụ: `storages/logs/app.log`     |
| `RequestLogPath` | string | Request log file path | Ví dụ: `storages/logs/request.log` |
| `EnableCaller`   | bool   | Show file:line        | `true`, `false`                    |
| `PrettyPrint`    | bool   | Pretty print console  | `true`, `false`                    |
| `DailyRotation`  | bool   | Enable daily rotation | `true`, `false`                    |

## Daily Rotation

Logger hỗ trợ daily rotation tự động cho log files. Khi bật `DailyRotation: true`, log files sẽ được tạo với format `filename-YYYY-MM-DD.log`.

### Ví dụ Daily Rotation

```go
// Cấu hình với daily rotation
config := logger.Config{
    Level:        "debug",
    Output:       "console,file",
    FilePath:     "storages/logs/app.log",
    RequestLogPath: "storages/logs/request.log",
    DailyRotation: true,
}

// Khởi tạo logger
logger.Init(config)

// Logs sẽ được ghi vào:
// - storages/logs/app-2024-01-15.log
// - storages/logs/request-2024-01-15.log
```

### File Structure với Daily Rotation

```
storages/logs/
├── app-2024-01-15.log      # Application logs
├── app-2024-01-16.log      # Application logs (next day)
├── request-2024-01-15.log  # Request logs
├── request-2024-01-16.log  # Request logs (next day)
└── cleanup-logs-2024-01-15.log  # Job-specific logs
```

## Sử Dụng

### Basic Logging

```go
import "api-core/pkg/logger"

// Info
logger.Info("Application started")
logger.Infof("Server running on port %d", 3000)

// Debug
logger.Debug("Debug information")
logger.Debugf("User ID: %s", userID)

// Warning
logger.Warn("This is a warning")
logger.Warnf("Rate limit exceeded: %d requests", count)

// Error
logger.Error("An error occurred")
logger.Errorf("Failed to connect to database: %v", err)
logger.ErrorWithErr(err, "Database connection failed")

// Fatal (exits program)
logger.Fatal("Critical error - shutting down")
logger.Fatalf("Cannot start server: %v", err)
```

### Logging với Fields

```go
// Single field
log := logger.WithField("user_id", "123")
log.Info().Msg("User logged in")

// Multiple fields
log := logger.WithFields(map[string]interface{}{
    "user_id": "123",
    "email":   "user@example.com",
    "role":    "admin",
})
log.Info().Msg("User action performed")
```

### Middleware

Có 3 loại middleware:

#### 1. SimpleMiddleware (không log body)

Middleware nhẹ nhất, chỉ log thông tin cơ bản. **Khuyến nghị cho production**.

```go
import (
    "api-core/pkg/logger"
    "github.com/go-chi/chi/v5"
)

r := chi.NewRouter()
r.Use(logger.SimpleMiddleware())
```

Output:

```json
{
  "level": "info",
  "request_id": "xyz123",
  "method": "GET",
  "uri": "/api/v1/users",
  "path": "/api/v1/users",
  "remote_addr": "127.0.0.1:12345",
  "status": 200,
  "duration_ms": 15,
  "time": "2025-10-21T14:30:00+07:00",
  "message": "Request completed"
}
```

#### 2. Middleware (log đầy đủ request/response)

Log đầy đủ request body, response body, và headers. **Khuyến nghị cho development**.

```go
r.Use(logger.Middleware())
```

Output bao gồm thêm:

```json
{
  "level": "info",
  "request_id": "xyz123",
  "method": "POST",
  "uri": "/api/v1/users",
  "path": "/api/v1/users",
  "remote_addr": "127.0.0.1:12345",
  "user_agent": "Mozilla/5.0...",
  "content_type": "application/json",
  "accept": "application/json",
  "status": 201,
  "duration_ms": 25,
  "request_body": "{\"name\":\"John\",\"email\":\"john@example.com\"}",
  "request_size": 52,
  "response_body": "{\"id\":\"123\",\"name\":\"John\",\"email\":\"john@example.com\"}",
  "response_size": 65,
  "response_content_type": "application/json",
  "message": "Request completed"
}
```

#### 3. MiddlewareWithConfig (tùy chỉnh)

Middleware có thể config linh hoạt theo nhu cầu.

```go
// Config tùy chỉnh
config := logger.MiddlewareConfig{
    LogRequestBody:  true,    // Log request body
    LogResponseBody: true,    // Log response body
    LogHeaders:      false,   // Không log headers
    MaxBodySize:     5000,    // Max 5KB body size
}

r.Use(logger.MiddlewareWithConfig(config))

// Hoặc dùng config mặc định
r.Use(logger.MiddlewareWithConfig(logger.DefaultMiddlewareConfig))
```

**Config Options:**

| Field             | Type | Default | Description                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `LogRequestBody`  | bool | `true`  | Log request body             |
| `LogResponseBody` | bool | `true`  | Log response body            |
| `LogHeaders`      | bool | `true`  | Log request headers          |
| `MaxBodySize`     | int  | `10000` | Max body size to log (bytes) |

⚠️ **Lưu ý**:

- Full middleware có thể ảnh hưởng performance vì phải buffer response body
- Body lớn hơn `MaxBodySize` sẽ không được log (chỉ log size)
- Sử dụng `SimpleMiddleware()` cho production để performance tốt nhất

### Logging trong Handlers

```go
import "api-core/pkg/logger"

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Log basic
    logger.RequestLogger(r, "Fetching user")

    // Log với fields
    logger.RequestLoggerWithFields(r, "User action", map[string]interface{}{
        "user_id": userID,
        "action":  "fetch",
    })

    // Log error
    if err != nil {
        logger.ErrorLogger(r, err, "Failed to fetch user")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}
```

## Log Format

### Console Format (Pretty Print = true)

```
2025-10-21 14:30:00 INF Starting ApiCore application...
2025-10-21 14:30:00 INF Dependencies initialized successfully
2025-10-21 14:30:00 INF Server starting on :3000
2025-10-21 14:30:05 INF Request completed request_id=abc123 method=GET path=/api/v1/users status=200 duration=15ms
```

### JSON Format (Pretty Print = false hoặc file)

```json
{
  "level": "info",
  "time": "2025-10-21T14:30:00+07:00",
  "message": "Request completed",
  "request_id": "abc123",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "duration_ms": 15
}
```

## Best Practices

### 1. Chọn Log Level Phù Hợp

```go
// Development
logger.Init(logger.Config{
    Level:       "debug",
    Output:      "both",
    PrettyPrint: true,
})

// Production
logger.Init(logger.Config{
    Level:       "info",
    Output:      "file",
    PrettyPrint: false,
})
```

### 2. Sử Dụng Fields cho Structured Logging

```go
// ❌ Không nên
logger.Infof("User %s created order %s with total %f", userID, orderID, total)

// ✅ Nên
logger.WithFields(map[string]interface{}{
    "user_id":  userID,
    "order_id": orderID,
    "total":    total,
}).Info().Msg("Order created")
```

### 3. Log Context trong Requests

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // Log start
    logger.RequestLogger(r, "Creating order")

    // Business logic...

    // Log success
    logger.RequestLoggerWithFields(r, "Order created", map[string]interface{}{
        "order_id": order.ID,
        "total":    order.Total,
    })

    // Log error
    if err != nil {
        logger.ErrorLogger(r, err, "Failed to create order")
        return
    }
}
```

### 4. Không Log Sensitive Data

```go
// ❌ Không nên log passwords, tokens
logger.WithFields(map[string]interface{}{
    "username": user.Username,
    "password": user.Password, // ❌ NEVER
}).Info().Msg("User login")

// ✅ Nên
logger.WithFields(map[string]interface{}{
    "username": user.Username,
    "user_id":  user.ID,
}).Info().Msg("User login successful")
```

## Log Rotation

Để rotate logs, có thể sử dụng các công cụ như:

### Linux (logrotate)

Tạo file `/etc/logrotate.d/apicore`:

```
/path/to/storages/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 www-data www-data
    sharedscripts
    postrotate
        systemctl reload apicore
    endscript
}
```

### Sử dụng lumberjack

```go
import "gopkg.in/natefinch/lumberjack.v2"

fileWriter := &lumberjack.Logger{
    Filename:   "storages/logs/app.log",
    MaxSize:    100, // MB
    MaxBackups: 3,
    MaxAge:     28, // days
    Compress:   true,
}
```

## Performance

### Benchmarks

- Simple logging: ~100ns/op
- Structured logging với fields: ~500ns/op
- Middleware (simple): adds ~10μs per request
- Middleware (full): adds ~50μs per request

### Tips

1. Sử dụng `SimpleMiddleware()` cho production
2. Chỉ log request/response body khi debug
3. Set log level = `info` hoặc `warn` cho production
4. Sử dụng async logging nếu cần performance cao hơn

## Troubleshooting

### Log file không được tạo

- Kiểm tra permissions của directory
- Đảm bảo path đúng

```go
// Fix: Tạo directory trước
os.MkdirAll("storages/logs", 0755)
```

### Logs không hiển thị màu

- Đảm bảo `PrettyPrint: true`
- Kiểm tra terminal có hỗ trợ ANSI colors

### Log quá nhiều

- Tăng log level lên `info` hoặc `warn`
- Sử dụng `SimpleMiddleware()` thay vì `Middleware()`
- Disable caller với `EnableCaller: false`

## Examples

Xem ví dụ đầy đủ tại:

- `cmd/app/main.go` - Khởi tạo logger
- `internal/app/user/controller.go` - Sử dụng logger trong handlers
- `pkg/logger/middleware.go` - Middleware implementation
