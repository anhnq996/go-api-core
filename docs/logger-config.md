# Logger Configuration

Logger configuration được quản lý thông qua environment variables để dễ dàng cấu hình cho các môi trường khác nhau.

## Environment Variables

### LOG_LEVEL

- **Mô tả**: Mức độ log (debug, info, warn, error)
- **Mặc định**: `debug`
- **Giá trị hợp lệ**: `debug`, `info`, `warn`, `error`

### LOG_OUTPUT

- **Mô tả**: Nơi xuất logs (có thể kết hợp nhiều)
- **Mặc định**: `console,file`
- **Giá trị hợp lệ**: `console`, `file`, `loki`
- **Ví dụ**: `console,file,loki`

### LOG_PATH

- **Mô tả**: Đường dẫn thư mục chứa logs
- **Mặc định**: `storages/logs`

### LOG_LOKI_URL

- **Mô tả**: URL của Loki server (bắt buộc nếu sử dụng loki output)
- **Mặc định**: `http://localhost:3100`

### LOG_ENABLE_CALLER

- **Mô tả**: Hiển thị file:line trong logs
- **Mặc định**: `false`
- **Giá trị hợp lệ**: `true`, `false`

### LOG_PRETTY_PRINT

- **Mô tả**: Format đẹp cho console output
- **Mặc định**: `true`
- **Giá trị hợp lệ**: `true`, `false`

### LOG_DAILY_ROTATION

- **Mô tả**: Bật daily rotation cho file logs
- **Mặc định**: `true`
- **Giá trị hợp lệ**: `true`, `false`

## Ví dụ Configuration

### Development Environment

```env
LOG_LEVEL=debug
LOG_OUTPUT=console,file
LOG_PATH=storages/logs
LOG_ENABLE_CALLER=true
LOG_PRETTY_PRINT=true
LOG_DAILY_ROTATION=false
```

### Production Environment

```env
LOG_LEVEL=info
LOG_OUTPUT=file,loki
LOG_PATH=/var/log/apicore
LOG_LOKI_URL=http://loki:3100
LOG_ENABLE_CALLER=false
LOG_PRETTY_PRINT=false
LOG_DAILY_ROTATION=true
```

### Testing Environment

```env
LOG_LEVEL=warn
LOG_OUTPUT=console
LOG_PRETTY_PRINT=true
LOG_DAILY_ROTATION=false
```

## File Structure với Daily Rotation

Khi `LOG_DAILY_ROTATION=true`, files sẽ được tạo với format:

```
storages/logs/
├── app-2025-10-24.log      # Application logs
├── app-2025-10-25.log      # Application logs (next day)
├── request-2025-10-24.log  # Request logs
├── request-2025-10-25.log  # Request logs (next day)
└── cleanup-logs-2025-10-24.log  # Job-specific logs
```

## Loki Integration

Khi sử dụng `loki` output, logs sẽ được gửi đến Loki server với các labels:

- **Application logs**: `job="apicore"`
- **Request logs**: `job="request"`
- **Job logs**: `job="job-name"`

## Validation

Config sẽ được validate khi khởi tạo logger:

- Kiểm tra level hợp lệ
- Kiểm tra output hợp lệ
- Kiểm tra Loki URL nếu sử dụng loki output
- Kiểm tra file paths có thể tạo được không

## Usage

```go
// Load config từ environment variables
loggerConfig := config.LoadLoggerConfig()

// Validate config
if err := loggerConfig.Validate(); err != nil {
    panic(fmt.Sprintf("Invalid logger config: %v", err))
}

// Initialize logger
if err := logger.Init(loggerConfig.ToLoggerConfig()); err != nil {
    panic(err)
}
```
