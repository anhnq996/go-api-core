# Cron Package

A distributed cron job scheduler with lock mechanism for multi-container deployments.

## Features

- **Distributed Locking**: Prevents duplicate job execution across multiple containers
- **Multiple Lock Backends**: Redis and in-memory lock managers
- **Retry Mechanism**: Configurable retry count and delay
- **Job Status Tracking**: Monitor job execution status and statistics
- **Flexible Scheduling**: Standard cron expressions
- **Timeout Support**: Configurable job timeouts
- **Metrics Ready**: Built-in support for metrics collection

## Installation

```bash
go get github.com/robfig/cron/v3
go get github.com/go-redis/redis/v8
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "your-project/pkg/cron"
    "github.com/go-redis/redis/v8"
)

// Define a simple job
type MyJob struct {
    name string
}

func (j *MyJob) Name() string {
    return j.name
}

func (j *MyJob) Schedule() string {
    return "*/5 * * * *" // Every 5 minutes
}

func (j *MyJob) Run(ctx context.Context) error {
    fmt.Printf("Running job: %s\n", j.name)
    return nil
}

func (j *MyJob) Timeout() time.Duration {
    return 5 * time.Minute
}

func (j *MyJob) RetryCount() int {
    return 3
}

func (j *MyJob) RetryDelay() time.Duration {
    return 10 * time.Second
}

func main() {
    // Create Redis client
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Create lock manager
    lockManager := cron.NewRedisLockManager(rdb, "myapp:cron:")

    // Create scheduler
    config := cron.Config{
        TimeZone: "UTC",
        LockTTL: 5 * time.Minute,
    }
    scheduler := cron.NewScheduler(lockManager, config)

    // Add job
    job := &MyJob{name: "cleanup"}
    scheduler.AddJob(job)

    // Start scheduler
    ctx := context.Background()
    scheduler.Start(ctx)

    // Keep running
    select {}
}
```

### Multi-Container Setup

```go
// In each container
func main() {
    // Use Redis for distributed locking
    rdb := redis.NewClient(&redis.Options{
        Addr: "redis-cluster:6379",
    })

    lockManager := cron.NewRedisLockManager(rdb, "myapp:cron:")

    config := cron.Config{
        TimeZone: "UTC",
        LockTTL: 10 * time.Minute, // Longer TTL for distributed setup
        MaxLockRetries: 5,
        LockRetryDelay: 2 * time.Second,
    }

    scheduler := cron.NewScheduler(lockManager, config)

    // Add jobs
    scheduler.AddJob(&CleanupJob{})
    scheduler.AddJob(&BackupJob{})
    scheduler.AddJob(&ReportJob{})

    // Start scheduler
    ctx := context.Background()
    scheduler.Start(ctx)

    // Graceful shutdown
    defer scheduler.Stop()

    // Keep running
    select {}
}
```

## Configuration

### Scheduler Config

```go
type Config struct {
    TimeZone         string        // Timezone for cron expressions
    LockTTL          time.Duration // Default lock TTL
    LockRetryDelay   time.Duration // Delay between lock retries
    MaxLockRetries   int           // Maximum lock acquisition retries
    JobTimeout       time.Duration // Default job timeout
    EnableMetrics    bool          // Enable metrics collection
    MetricsPrefix    string        // Metrics prefix
}
```

### Job Interface

```go
type Job interface {
    Name() string                    // Unique job name
    Schedule() string                // Cron expression
    Run(ctx context.Context) error   // Job execution
    Timeout() time.Duration          // Job timeout
    RetryCount() int                 // Number of retries
    RetryDelay() time.Duration       // Delay between retries
}
```

## Lock Managers

### Redis Lock Manager

```go
// For distributed deployments
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
lockManager := cron.NewRedisLockManager(rdb, "myapp:cron:")
```

### Memory Lock Manager

```go
// For single instance deployments
lockManager := cron.NewMemoryLockManager()
```

## Job Status Monitoring

```go
// Get status of all jobs
statuses := scheduler.GetJobStatuses()
for name, status := range statuses {
    fmt.Printf("Job: %s, Last Run: %s, Next Run: %s, Is Running: %v\n",
        name, status.LastRun, status.NextRun, status.IsRunning)
}

// Get status of specific job
status, err := scheduler.GetJobStatus("cleanup")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Cleanup job status: %+v\n", status)
```

## Advanced Usage

### Custom Job with Error Handling

```go
type DatabaseCleanupJob struct {
    db *sql.DB
}

func (j *DatabaseCleanupJob) Name() string {
    return "database_cleanup"
}

func (j *DatabaseCleanupJob) Schedule() string {
    return "0 2 * * *" // Daily at 2 AM
}

func (j *DatabaseCleanupJob) Run(ctx context.Context) error {
    // Check if context is cancelled
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Perform cleanup
    _, err := j.db.ExecContext(ctx, "DELETE FROM logs WHERE created_at < ?",
        time.Now().Add(-30*24*time.Hour))
    if err != nil {
        return fmt.Errorf("failed to cleanup logs: %w", err)
    }

    return nil
}

func (j *DatabaseCleanupJob) Timeout() time.Duration {
    return 30 * time.Minute
}

func (j *DatabaseCleanupJob) RetryCount() int {
    return 2
}

func (j *DatabaseCleanupJob) RetryDelay() time.Duration {
    return 5 * time.Minute
}
```

### Job with Metrics

```go
type MetricsJob struct {
    counter prometheus.Counter
}

func (j *MetricsJob) Run(ctx context.Context) error {
    // Increment counter
    j.counter.Inc()

    // Your job logic here
    return nil
}
```

## Best Practices

1. **Use Redis Lock Manager** for multi-container deployments
2. **Set appropriate timeouts** for your jobs
3. **Handle context cancellation** in long-running jobs
4. **Use meaningful job names** for monitoring
5. **Set reasonable retry counts** to avoid infinite loops
6. **Monitor job status** for debugging and alerting
7. **Use timezone-aware scheduling** for global applications

## Error Handling

The scheduler automatically handles:

- Lock acquisition failures
- Job execution timeouts
- Retry logic with exponential backoff
- Context cancellation
- Panic recovery

## Monitoring

Track job execution with:

- Job status API
- Metrics collection (if enabled)
- Logging integration
- Health checks

## Examples

See the `examples/` directory for complete examples:

- Basic job scheduling
- Multi-container setup
- Custom job implementations
- Error handling patterns
- Monitoring and metrics
