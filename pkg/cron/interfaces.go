package cron

import (
	"context"
	"time"
)

// Job represents a cron job
type Job interface {
	// Name returns the unique name of the job
	Name() string

	// Schedule returns the cron schedule expression
	Schedule() string

	// Run executes the job
	Run(ctx context.Context) error

	// Timeout returns the maximum execution time for the job
	Timeout() time.Duration

	// RetryCount returns the number of retries on failure
	RetryCount() int

	// RetryDelay returns the delay between retries
	RetryDelay() time.Duration
}

// LockManager handles distributed locking for cron jobs
type LockManager interface {
	// AcquireLock attempts to acquire a lock for the given job
	AcquireLock(ctx context.Context, jobName string, ttl time.Duration) (bool, error)

	// ReleaseLock releases the lock for the given job
	ReleaseLock(ctx context.Context, jobName string) error

	// ExtendLock extends the lock TTL for the given job
	ExtendLock(ctx context.Context, jobName string, ttl time.Duration) error

	// IsLocked checks if a job is currently locked
	IsLocked(ctx context.Context, jobName string) (bool, error)
}

// Scheduler manages and executes cron jobs
type Scheduler interface {
	// AddJob adds a job to the scheduler
	AddJob(job Job) error

	// RemoveJob removes a job from the scheduler
	RemoveJob(jobName string) error

	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop stops the scheduler
	Stop() error

	// IsRunning returns true if the scheduler is running
	IsRunning() bool

	// GetJobStatus returns the status of a specific job
	GetJobStatus(jobName string) (*JobStatus, error)

	// GetJobStatuses returns the status of all jobs
	GetJobStatuses() map[string]*JobStatus
}

// JobStatus represents the status of a cron job
type JobStatus struct {
	Name         string    `json:"name"`
	Schedule     string    `json:"schedule"`
	LastRun      time.Time `json:"last_run"`
	NextRun      time.Time `json:"next_run"`
	IsRunning    bool      `json:"is_running"`
	IsLocked     bool      `json:"is_locked"`
	RunCount     int64     `json:"run_count"`
	SuccessCount int64     `json:"success_count"`
	ErrorCount   int64     `json:"error_count"`
	LastError    string    `json:"last_error,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// JobResult represents the result of a job execution
type JobResult struct {
	JobName    string        `json:"job_name"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	RetryCount int           `json:"retry_count"`
}

// Config represents the configuration for the cron scheduler
type Config struct {
	// TimeZone specifies the timezone for the scheduler
	TimeZone string `json:"time_zone"`

	// LockTTL specifies the default lock TTL for jobs
	LockTTL time.Duration `json:"lock_ttl"`

	// LockRetryDelay specifies the delay between lock acquisition retries
	LockRetryDelay time.Duration `json:"lock_retry_delay"`

	// MaxLockRetries specifies the maximum number of lock acquisition retries
	MaxLockRetries int `json:"max_lock_retries"`

	// JobTimeout specifies the default timeout for jobs
	JobTimeout time.Duration `json:"job_timeout"`

	// EnableMetrics specifies whether to enable metrics collection
	EnableMetrics bool `json:"enable_metrics"`

	// MetricsPrefix specifies the prefix for metrics
	MetricsPrefix string `json:"metrics_prefix"`
}
