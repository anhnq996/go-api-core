# Loki & Grafana Setup Guide

Hướng dẫn setup Loki và Grafana để xem logs từ ApiCore.

## Tổng Quan

- **Loki**: Log aggregation system (như Elasticsearch nhưng nhẹ hơn)
- **Grafana**: Visualization platform để xem và query logs
- **ApiCore**: Gửi logs trực tiếp đến Loki

## Kiến Trúc

```
ApiCore (Go App)
    ↓ (HTTP push)
Loki (Port 3100)
    ↓ (query)
Grafana (Port 3001)
    ↓ (view)
Browser
```

## Quick Start

### 1. Start Loki & Grafana

```bash
# Start services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f loki
docker-compose logs -f grafana
```

### 2. Start ApiCore

```bash
go run cmd/app/main.go
```

ApiCore sẽ tự động gửi logs đến:

- **Console**: Terminal output (pretty print)
- **File**: `storages/logs/app.log`
- **Loki**: `http://localhost:3100`

### 3. Access Grafana

Mở browser: **http://localhost:3001**

- Username: `admin`
- Password: `admin`

Loki datasource đã được cấu hình tự động!

## Xem Logs trong Grafana

### Bước 1: Mở Explore

1. Click icon **Explore** (🧭) ở sidebar trái
2. Chọn datasource: **Loki**

### Bước 2: Query Logs

#### Query tất cả logs

```logql
{job="apicore"}
```

#### Query theo log level

```logql
# Chỉ errors
{job="apicore"} |= "level=error"

# Chỉ warnings
{job="apicore"} |= "level=warn"

# Info và above
{job="apicore"} |= `level="info"`
```

#### Query theo endpoint

```logql
# Logs từ user endpoint
{job="apicore"} |= "/api/v1/users"

# POST requests only
{job="apicore"} |= "method=POST"
```

#### Query theo request ID

```logql
{job="apicore"} |= "request_id=RSHCM207PCC/wVCCJDOtcR-000001"
```

#### Query với filters

```logql
# Status 5xx errors
{job="apicore"} |= "status=5"

# Slow requests (>100ms)
{job="apicore"} | json | duration_ms > 100

# Failed operations
{job="apicore"} |= "Failed"
```

### Bước 3: Tạo Dashboard

1. Click **Dashboards** > **New Dashboard**
2. Add Panel > Chọn visualization type
3. Query logs với LogQL
4. Save dashboard

#### Dashboard Examples

**Panel 1: Request Rate**

```logql
rate({job="apicore"}[5m])
```

**Panel 2: Error Rate**

```logql
rate({job="apicore"} |= "level=error"[5m])
```

**Panel 3: Response Time (p95)**

```logql
quantile_over_time(0.95, {job="apicore"} | json | unwrap duration_ms [5m])
```

**Panel 4: Status Codes**

```logql
sum by (status) (rate({job="apicore"} | json [5m]))
```

## LogQL Cheatsheet

### Basic Queries

```logql
# All logs
{job="apicore"}

# Filter by string
{job="apicore"} |= "error"

# Regex filter
{job="apicore"} |~ "error|failed"

# NOT filter
{job="apicore"} != "health"
```

### JSON Parsing

```logql
# Parse JSON and filter
{job="apicore"} | json | status >= 400

# Extract fields
{job="apicore"} | json | line_format "{{.method}} {{.path}} {{.status}}"
```

### Aggregations

```logql
# Count logs
count_over_time({job="apicore"}[5m])

# Sum
sum(rate({job="apicore"}[5m]))

# Average
avg(rate({job="apicore"} | json | unwrap duration_ms [5m]))
```

## Configuration

### Logger Config

File: `cmd/app/main.go`

```go
logger.Init(logger.Config{
    Level:        "debug",              // log level
    Output:       "console,file,loki",  // multiple outputs
    FilePath:     "storages/logs/app.log",
    LokiURL:      "http://localhost:3100",
    EnableCaller: false,
    PrettyPrint:  true,
})
```

**Output Options:**

- `console` - Terminal only
- `file` - File only
- `loki` - Loki only
- `console,file` - Console + File
- `console,loki` - Console + Loki
- `file,loki` - File + Loki
- `console,file,loki` - All three

### Loki Config

File: `config/loki-config.yaml`

Các settings quan trọng:

```yaml
limits_config:
  ingestion_rate_mb: 16 # Max ingestion rate
  per_stream_rate_limit: 8MB # Per stream limit

compactor:
  retention_enabled: true # Enable log retention
  retention_delete_delay: 2h # Delete after
```

## Troubleshooting

### Loki không nhận logs

1. Check Loki status:

```bash
curl http://localhost:3100/ready
```

2. Check logs:

```bash
docker-compose logs loki
```

3. Kiểm tra ApiCore có gửi logs không:

```bash
# Should see logs trong terminal và file
go run cmd/app/main.go
```

### Grafana không kết nối được Loki

1. Check datasource config:

```bash
cat config/grafana-datasources.yaml
```

2. Trong Grafana: Configuration > Data sources > Loki > Test

3. Đảm bảo Loki URL đúng: `http://loki:3100`

### Logs không hiển thị trong Grafana

1. Check time range (góc phải trên)
2. Try query: `{job="apicore"}`
3. Click **Live** để xem real-time logs

### Performance Issues

**Giảm log volume:**

```go
// Chỉ log console (fastest)
Output: "console",

// Hoặc chỉ Loki
Output: "loki",
```

**Tăng batch size:**
Trong `pkg/logger/logger.go`:

```go
cfg.BatchWait = 5 * time.Second
cfg.BatchSize = 500 * 1024 // 500KB
```

## Best Practices

### 1. Sử dụng Labels Đúng Cách

```go
// ❌ Không nên - quá nhiều labels
logger.WithFields(map[string]interface{}{
    "user_id": "123",
    "ip": "1.2.3.4",
    "endpoint": "/api/users",
})

// ✅ Nên - chỉ dùng indexed fields
logger.RequestLoggerWithFields(r, "User action", map[string]interface{}{
    "action": "create",
    "resource": "user",
})
```

### 2. Query Optimization

```logql
# ❌ Slow - scan all logs
{job="apicore"} | json | status == 500

# ✅ Fast - filter first
{job="apicore"} |= "status=500" | json
```

### 3. Retention Policy

Set retention theo nhu cầu:

```yaml
# Short retention (7 days)
compactor:
  retention_enabled: true
  retention_delete_delay: 168h

# Long retention (30 days)
retention_delete_delay: 720h
```

## Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart service
docker-compose restart loki

# Remove all data (including logs)
docker-compose down -v

# Check resource usage
docker stats apicore-loki apicore-grafana
```

## URLs

- **Grafana**: http://localhost:3001
- **Loki API**: http://localhost:3100
- **Loki Health**: http://localhost:3100/ready
- **Loki Metrics**: http://localhost:3100/metrics

## Next Steps

1. ✅ Setup alerts trong Grafana
2. ✅ Tạo dashboards cho monitoring
3. ✅ Configure retention policies
4. ✅ Add more log labels nếu cần
5. ✅ Integrate với Prometheus (optional)

## Resources

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Language](https://grafana.com/docs/loki/latest/logql/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
