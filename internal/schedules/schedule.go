package schedules

import (
	"context"
	"fmt"
	"log"
	"time"

	"api-core/internal/schedules/jobs"
	"api-core/pkg/cron"
)

// JobWrapper wraps jobs.Job to implement cron.Job interface
type JobWrapper struct {
	job      jobs.Job
	schedule string
}

func (jw *JobWrapper) Name() string {
	return jw.job.Name()
}

func (jw *JobWrapper) Schedule() string {
	return jw.schedule
}

func (jw *JobWrapper) Run(ctx context.Context) error {
	return jw.job.Run(ctx)
}

func (jw *JobWrapper) Timeout() time.Duration {
	return jw.job.Timeout()
}

func (jw *JobWrapper) RetryCount() int {
	return jw.job.RetryCount()
}

func (jw *JobWrapper) RetryDelay() time.Duration {
	return jw.job.RetryDelay()
}

// JobConfig cấu hình cho job
type JobConfig struct {
	Name     string
	Schedule string
	Job      cron.Job
}

// ScheduleManager quản lý tất cả cron jobs
type ScheduleManager struct {
	scheduler   cron.Scheduler
	lockManager cron.LockManager
}

// NewScheduleManager tạo schedule manager mới
func NewScheduleManager(lockManager cron.LockManager) *ScheduleManager {
	config := cron.Config{
		TimeZone:       "UTC",
		LockTTL:        5 * time.Minute,
		LockRetryDelay: 1 * time.Second,
		MaxLockRetries: 3,
		JobTimeout:     1 * time.Minute,
		EnableMetrics:  true,
		MetricsPrefix:  "api_core",
	}

	scheduler := cron.NewScheduler(lockManager, config)

	return &ScheduleManager{
		scheduler:   scheduler,
		lockManager: lockManager,
	}
}

// RegisterAllJobs đăng ký tất cả jobs
func (sm *ScheduleManager) RegisterAllJobs() error {
	// Cron expression cho các jobs
	jobCron := map[string]string{
		"cleanup-logs":       "0 0 * * *", // Mỗi ngày lúc 0h
		"cleanup-temp-files": "0 0 * * *", // Mỗi ngày lúc 0h
		"health-check":       "0 * * * *", // Mỗi giờ
	}

	// Đăng ký các jobs
	jobsToRegister := []JobConfig{
		{
			Name:     "cleanup-logs",
			Schedule: jobCron["cleanup-logs"], // Mỗi phút
			Job:      &JobWrapper{job: &jobs.CleanupLogsJob{}, schedule: jobCron["cleanup-logs"]},
		},
		{
			Name:     "cleanup-temp-files",
			Schedule: jobCron["cleanup-temp-files"], // Mỗi 2 phút
			Job:      &JobWrapper{job: &jobs.CleanupTempFilesJob{}, schedule: jobCron["cleanup-temp-files"]},
		},
		{
			Name:     "health-check",
			Schedule: jobCron["health-check"], // Mỗi 10 phút
			Job:      &JobWrapper{job: &jobs.HealthCheckJob{}, schedule: jobCron["health-check"]},
		},
	}

	// Đăng ký từng job
	for _, jobConfig := range jobsToRegister {
		if err := sm.scheduler.AddJob(jobConfig.Job); err != nil {
			return fmt.Errorf("failed to register job %s: %w", jobConfig.Name, err)
		}
		log.Printf("Registered job: %s with schedule: %s", jobConfig.Name, jobConfig.Schedule)
	}

	return nil
}

// Start bắt đầu scheduler
func (sm *ScheduleManager) Start(ctx context.Context) error {
	log.Println("Starting schedule manager...")

	if err := sm.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	log.Println("Schedule manager started successfully")

	// Log job statuses
	statuses := sm.scheduler.GetJobStatuses()
	for name, status := range statuses {
		log.Printf("Job %s: %s - Running: %v", name, status.Schedule, status.IsRunning)
	}

	return nil
}

// Stop dừng scheduler
func (sm *ScheduleManager) Stop() error {
	if err := sm.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	log.Println("Schedule manager stopped")
	return nil
}

// GetJobStatuses lấy trạng thái tất cả jobs
func (sm *ScheduleManager) GetJobStatuses() map[string]*cron.JobStatus {
	return sm.scheduler.GetJobStatuses()
}

// GetJobStatus lấy trạng thái job cụ thể
func (sm *ScheduleManager) GetJobStatus(jobName string) (*cron.JobStatus, error) {
	return sm.scheduler.GetJobStatus(jobName)
}

// IsRunning kiểm tra scheduler có đang chạy không
func (sm *ScheduleManager) IsRunning() bool {
	return sm.scheduler.IsRunning()
}

// InitScheduleManager khởi tạo schedule manager với logger
func InitScheduleManager(lockManager cron.LockManager) (*ScheduleManager, error) {
	// Schedule manager sử dụng logger đã được khởi tạo từ main
	// Không cần khởi tạo lại logger ở đây để tránh ghi đè RequestLogger

	// Tạo schedule manager
	manager := NewScheduleManager(lockManager)

	// Đăng ký tất cả jobs
	if err := manager.RegisterAllJobs(); err != nil {
		return nil, fmt.Errorf("failed to register jobs: %w", err)
	}

	return manager, nil
}
