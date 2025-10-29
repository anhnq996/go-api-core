# Schedules Module

Module quản lý tất cả cron jobs trong ứng dụng API Core.

## Cấu trúc thư mục

```
internal/schedules/
├── schedule.go         # Schedule manager chính
├── example.go          # Ví dụ sử dụng
├── README.md           # Hướng dẫn này
└── jobs/               # Thư mục chứa các job
    ├── job_interface.go # Interface cho jobs
    ├── cleanup_logs.go
    ├── backup_database.go
    ├── send_notifications.go
    ├── cleanup_temp_files.go
    ├── health_check.go
    └── generate_reports.go
```

## Cách sử dụng

### 1. Khởi tạo Schedule Manager

```go
package main

import (
    "context"
    "log"

    "api-core/internal/schedules"
    "api-core/pkg/cron"
    "github.com/go-redis/redis/v8"
)

func main() {
    // Tạo Redis client
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Tạo lock manager
    lockManager := cron.NewRedisLockManager(rdb, "api-core:cron:")

    // Khởi tạo schedule manager
    manager, err := schedules.InitScheduleManager(lockManager)
    if err != nil {
        log.Fatalf("Failed to initialize schedule manager: %v", err)
    }

    // Start scheduler
    ctx := context.Background()
    if err := manager.Start(ctx); err != nil {
        log.Fatalf("Failed to start schedule manager: %v", err)
    }
    defer manager.Stop()

    // Keep running
    select {}
}
```

### 2. Sử dụng Memory Lock Manager (Single Instance)

```go
// Sử dụng memory lock manager cho single instance
lockManager := cron.NewMemoryLockManager()
manager, err := schedules.InitScheduleManager(lockManager)
```

### 3. Monitor Job Status

```go
// Lấy trạng thái tất cả jobs
statuses := manager.GetJobStatuses()
for name, status := range statuses {
    fmt.Printf("Job: %s, Last Run: %s, Run Count: %d\n",
        name, status.LastRun.Format("15:04:05"), status.RunCount)
}

// Lấy trạng thái job cụ thể
status, err := manager.GetJobStatus("cleanup-logs")
if err != nil {
    log.Printf("Failed to get job status: %v", err)
}
```

## Danh sách Jobs

### 1. Cleanup Logs Job

- **File**: `jobs/cleanup_logs.go`
- **Schedule**: `0 2 * * *` (Mỗi ngày lúc 02:00)
- **Mô tả**: Xóa log files cũ hơn 30 ngày
- **Timeout**: 10 phút
- **Retry**: 2 lần

### 2. Backup Database Job

- **File**: `jobs/backup_database.go`
- **Schedule**: `0 3 * * 0` (Mỗi Chủ nhật lúc 03:00)
- **Mô tả**: Backup database và xóa backup cũ hơn 7 ngày
- **Timeout**: 30 phút
- **Retry**: 1 lần

### 3. Send Notifications Job

- **File**: `jobs/send_notifications.go`
- **Schedule**: `*/5 * * * *` (Mỗi 5 phút)
- **Mô tả**: Gửi notifications pending
- **Timeout**: 5 phút
- **Retry**: 3 lần

### 4. Cleanup Temp Files Job

- **File**: `jobs/cleanup_temp_files.go`
- **Schedule**: `0 */6 * * *` (Mỗi 6 giờ)
- **Mô tả**: Xóa temp files cũ hơn 1 giờ
- **Timeout**: 15 phút
- **Retry**: 2 lần

### 5. Health Check Job

- **File**: `jobs/health_check.go`
- **Schedule**: `*/10 * * * *` (Mỗi 10 phút)
- **Mô tả**: Kiểm tra health của các services
- **Timeout**: 5 phút
- **Retry**: 3 lần

### 6. Generate Reports Job

- **File**: `jobs/generate_reports.go`
- **Schedule**: `0 1 * * 1` (Mỗi Thứ 2 lúc 01:00)
- **Mô tả**: Tạo reports định kỳ và xóa reports cũ
- **Timeout**: 20 phút
- **Retry**: 1 lần

## Thêm Job Mới

### 1. Tạo Job File

Tạo file mới trong thư mục `jobs/`, ví dụ `jobs/my_new_job.go`:

```go
package jobs

import (
    "context"
    "time"
    "api-core/pkg/logger"
)

type MyNewJob struct{}

func (j *MyNewJob) Name() string {
    return "my-new-job"
}

func (j *MyNewJob) Schedule() string {
    return "0 0 * * *" // Mỗi ngày lúc 00:00
}

func (j *MyNewJob) Run(ctx context.Context) error {
    jobLogger := logger.GetJobLogger(j.Name())
    jobLogger.Info().Msg("Running my new job")

    // Logic của job ở đây

    return nil
}

func (j *MyNewJob) Timeout() time.Duration {
    return 5 * time.Minute
}

func (j *MyNewJob) RetryCount() int {
    return 2
}

func (j *MyNewJob) RetryDelay() time.Duration {
    return 1 * time.Minute
}
```

### 2. Đăng ký Job

Thêm job vào `schedule.go` trong hàm `RegisterAllJobs()`:

```go
{
    name:     "my-new-job",
    schedule: "0 0 * * *", // Mỗi ngày lúc 00:00
    job:      &jobs.MyNewJob{},
},
```

## Cron Expression

Sử dụng cron expression chuẩn:

- `"* * * * *"` - Mỗi phút
- `"0 * * * *"` - Mỗi giờ
- `"0 0 * * *"` - Mỗi ngày lúc 00:00
- `"0 0 * * 0"` - Mỗi Chủ nhật lúc 00:00
- `"*/5 * * * *"` - Mỗi 5 phút
- `"0 0 1 * *"` - Ngày đầu tháng lúc 00:00

## Dynamic Logger

Mỗi job sẽ có logger riêng, log được ghi vào file riêng:

- `cleanup-logs.log`
- `backup-database.log`
- `send-notifications.log`
- `cleanup-temp-files.log`
- `health-check.log`
- `generate-reports.log`

## Chạy Example

```bash
# Chạy ví dụ
go run internal/schedules/example.go
```

## Production Usage

Trong production, bạn có thể tích hợp schedule manager vào main application:

```go
// Trong cmd/app/main.go
func main() {
    // ... other initialization

    // Initialize schedule manager
    lockManager := cron.NewRedisLockManager(redisClient, "api-core:cron:")
    scheduleManager, err := schedules.InitScheduleManager(lockManager)
    if err != nil {
        log.Fatalf("Failed to initialize schedule manager: %v", err)
    }

    // Start scheduler
    if err := scheduleManager.Start(ctx); err != nil {
        log.Fatalf("Failed to start schedule manager: %v", err)
    }

    // ... rest of application
}
```
