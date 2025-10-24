package jobs

import (
	"context"
	"time"
)

// Job interface cho các jobs - không cần Schedule() method nữa
type Job interface {
	// Name returns the unique name of the job
	Name() string

	// Run executes the job
	Run(ctx context.Context) error

	// Timeout returns the maximum execution time for the job
	Timeout() time.Duration

	// RetryCount returns the number of retries on failure
	RetryCount() int

	// RetryDelay returns the delay between retries
	RetryDelay() time.Duration
}
