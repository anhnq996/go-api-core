package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// SchedulerImpl implements the Scheduler interface
type SchedulerImpl struct {
	cron        *cron.Cron
	jobs        map[string]Job
	jobStatuses map[string]*JobStatus
	lockManager LockManager
	config      Config
	mu          sync.RWMutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewScheduler creates a new cron scheduler
func NewScheduler(lockManager LockManager, config Config) *SchedulerImpl {
	if config.TimeZone == "" {
		config.TimeZone = "UTC"
	}
	if config.LockTTL == 0 {
		config.LockTTL = 30 * time.Second // Giảm TTL xuống 30 giây
	}
	if config.LockRetryDelay == 0 {
		config.LockRetryDelay = 1 * time.Second
	}
	if config.MaxLockRetries == 0 {
		config.MaxLockRetries = 3
	}
	if config.JobTimeout == 0 {
		config.JobTimeout = 10 * time.Minute
	}
	if config.MetricsPrefix == "" {
		config.MetricsPrefix = "cron"
	}

	// Create cron scheduler with timezone
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		location = time.UTC
	}

	c := cron.New(
		cron.WithLocation(location),
		cron.WithLogger(cron.DefaultLogger),
	)

	return &SchedulerImpl{
		cron:        c,
		jobs:        make(map[string]Job),
		jobStatuses: make(map[string]*JobStatus),
		lockManager: lockManager,
		config:      config,
	}
}

// AddJob adds a job to the scheduler
func (s *SchedulerImpl) AddJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job already exists
	if _, exists := s.jobs[job.Name()]; exists {
		return fmt.Errorf("job %s already exists", job.Name())
	}

	// Add job to cron scheduler
	_, err := s.cron.AddFunc(job.Schedule(), s.createJobWrapper(job))
	if err != nil {
		return fmt.Errorf("failed to add job %s: %w", job.Name(), err)
	}

	// Store job and initialize status
	s.jobs[job.Name()] = job
	s.jobStatuses[job.Name()] = &JobStatus{
		Name:      job.Name(),
		Schedule:  job.Schedule(),
		CreatedAt: time.Now(),
	}

	// If scheduler is running, start the job immediately
	if s.running {
		s.cron.Start()
	}

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *SchedulerImpl) RemoveJob(jobName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job exists
	_, exists := s.jobs[jobName]
	if !exists {
		return fmt.Errorf("job %s does not exist", jobName)
	}

	// Remove from cron scheduler
	entries := s.cron.Entries()
	for _, entry := range entries {
		// We need to find the entry by comparing the job name
		// This is a limitation of the robfig/cron library
		// In a real implementation, you might want to track entry IDs
		_ = entry
	}

	// Remove from maps
	delete(s.jobs, jobName)
	delete(s.jobStatuses, jobName)

	// Release any existing lock
	_ = s.lockManager.ReleaseLock(context.Background(), jobName)

	return nil
}

// Start starts the scheduler
func (s *SchedulerImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.cron.Start()

	// Start cleanup goroutine for memory locks
	if _, ok := s.lockManager.(*MemoryLockManager); ok {
		go s.cleanupExpiredLocks()
	}

	return nil
}

// Stop stops the scheduler
func (s *SchedulerImpl) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	s.cron.Stop()
	if s.cancel != nil {
		s.cancel()
	}
	s.running = false

	return nil
}

// IsRunning returns true if the scheduler is running
func (s *SchedulerImpl) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetJobStatus returns the status of a specific job
func (s *SchedulerImpl) GetJobStatus(jobName string) (*JobStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.jobStatuses[jobName]
	if !exists {
		return nil, fmt.Errorf("job %s does not exist", jobName)
	}

	// Update next run time
	_ = s.jobs[jobName]
	entries := s.cron.Entries()
	for _, entry := range entries {
		// Find the entry for this job and update next run time
		// This is a simplified approach - in practice you'd track entry IDs
		status.NextRun = entry.Next
		break
	}

	// Check if job is locked
	isLocked, _ := s.lockManager.IsLocked(context.Background(), jobName)
	status.IsLocked = isLocked

	return status, nil
}

// GetJobStatuses returns the status of all jobs
func (s *SchedulerImpl) GetJobStatuses() map[string]*JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make(map[string]*JobStatus)
	for name, status := range s.jobStatuses {
		// Create a copy to avoid race conditions
		statusCopy := *status

		// Update next run time
		entries := s.cron.Entries()
		for _, entry := range entries {
			statusCopy.NextRun = entry.Next
			break
		}

		// Check if job is locked
		isLocked, _ := s.lockManager.IsLocked(context.Background(), name)
		statusCopy.IsLocked = isLocked

		statuses[name] = &statusCopy
	}

	return statuses
}

// createJobWrapper creates a wrapper function for a job that handles locking and execution
func (s *SchedulerImpl) createJobWrapper(job Job) func() {
	return func() {
		// Check if scheduler context is still valid
		if s.ctx == nil {
			fmt.Printf("Job %s: scheduler context is nil\n", job.Name())
			return
		}

		select {
		case <-s.ctx.Done():
			fmt.Printf("Job %s: scheduler context cancelled: %v\n", job.Name(), s.ctx.Err())
			return
		default:
		}

		ctx, cancel := context.WithTimeout(s.ctx, job.Timeout())
		defer cancel()

		fmt.Printf("Job %s: starting execution\n", job.Name())

		// Try to acquire lock
		acquired, err := s.acquireLockWithRetry(ctx, job.Name())
		if err != nil {
			fmt.Printf("Job %s: failed to acquire lock: %v\n", job.Name(), err)
			s.updateJobStatus(job.Name(), false, fmt.Sprintf("failed to acquire lock: %v", err))
			return
		}

		if !acquired {
			// Another instance is running this job
			fmt.Printf("Job %s: lock not acquired, another instance running\n", job.Name())
			return
		}

		// Execute job with retries
		s.executeJobWithRetry(ctx, job)
	}
}

// acquireLockWithRetry attempts to acquire a lock with retries
func (s *SchedulerImpl) acquireLockWithRetry(ctx context.Context, jobName string) (bool, error) {
	for i := 0; i < s.config.MaxLockRetries; i++ {
		acquired, err := s.lockManager.AcquireLock(ctx, jobName, s.config.LockTTL)
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(s.config.LockRetryDelay):
			continue
		}
	}

	return false, nil
}

// executeJobWithRetry executes a job with retries
func (s *SchedulerImpl) executeJobWithRetry(ctx context.Context, job Job) {
	var lastErr error
	retryCount := 0

	// Ensure lock is released when function exits
	defer func() {
		// Release lock after job completion
		if err := s.lockManager.ReleaseLock(ctx, job.Name()); err != nil {
			fmt.Printf("Job %s: failed to release lock: %v\n", job.Name(), err)
		} else {
			fmt.Printf("Job %s: lock released successfully\n", job.Name())
		}
	}()

	for retryCount <= job.RetryCount() {
		// Update job status
		s.updateJobStatus(job.Name(), true, "")

		// Execute job
		startTime := time.Now()
		err := job.Run(ctx)
		duration := time.Since(startTime)

		if err == nil {
			// Job succeeded
			s.updateJobStatus(job.Name(), false, "")
			s.recordJobResult(job.Name(), startTime, duration, true, "", retryCount)
			return
		}

		lastErr = err
		retryCount++

		// If we have retries left, wait before retrying
		if retryCount <= job.RetryCount() {
			select {
			case <-ctx.Done():
				s.updateJobStatus(job.Name(), false, fmt.Sprintf("job cancelled: %v", ctx.Err()))
				return
			case <-time.After(job.RetryDelay()):
				continue
			}
		}
	}

	// Job failed after all retries
	s.updateJobStatus(job.Name(), false, lastErr.Error())
	s.recordJobResult(job.Name(), time.Now(), 0, false, lastErr.Error(), retryCount-1)
}

// updateJobStatus updates the status of a job
func (s *SchedulerImpl) updateJobStatus(jobName string, isRunning bool, lastError string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.jobStatuses[jobName]
	if !exists {
		return
	}

	status.IsRunning = isRunning
	if lastError != "" {
		status.LastError = lastError
		status.ErrorCount++
	} else if !isRunning {
		status.SuccessCount++
	}

	if !isRunning {
		status.LastRun = time.Now()
		status.RunCount++
	}
}

// recordJobResult records the result of a job execution
func (s *SchedulerImpl) recordJobResult(jobName string, startTime time.Time, duration time.Duration, success bool, error string, retryCount int) {
	// This could be extended to store job results in a database or metrics system
	_ = JobResult{
		JobName:    jobName,
		StartTime:  startTime,
		EndTime:    startTime.Add(duration),
		Duration:   duration,
		Success:    success,
		Error:      error,
		RetryCount: retryCount,
	}
}

// cleanupExpiredLocks periodically cleans up expired locks for memory lock manager
func (s *SchedulerImpl) cleanupExpiredLocks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if memLockManager, ok := s.lockManager.(*MemoryLockManager); ok {
				memLockManager.CleanupExpiredLocks()
			}
		}
	}
}
