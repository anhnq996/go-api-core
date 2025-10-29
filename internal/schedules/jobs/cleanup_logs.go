package jobs

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"api-core/pkg/logger"
)

// CleanupLogsJob xóa log files cũ
type CleanupLogsJob struct{}

func (j *CleanupLogsJob) Name() string {
	return "cleanup-logs"
}

func (j *CleanupLogsJob) Run(ctx context.Context) error {
	jobLogger := logger.GetJobLogger(j.Name())
	jobLogger.Info().Msg("Starting cleanup logs job")

	// Đường dẫn thư mục logs
	logsDir := "storages/logs"

	// Xóa logs cũ hơn 30 ngày
	cutoffTime := time.Now().AddDate(0, 0, -30)

	deletedCount := 0
	err := filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Chỉ xử lý files, không phải directories
		if info.IsDir() {
			return nil
		}

		// Kiểm tra file có cũ hơn cutoff time không
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				jobLogger.Error().Err(err).Str("file", path).Msg("Failed to delete log file")
				return err
			}

			jobLogger.Info().Str("file", path).Msg("Deleted old log file")
			deletedCount++
		}

		return nil
	})

	if err != nil {
		jobLogger.Error().Err(err).Msg("Failed to cleanup logs")
		return err
	}

	jobLogger.Info().Int("deleted_count", deletedCount).Msg("Cleanup logs job completed")
	return nil
}

func (j *CleanupLogsJob) Timeout() time.Duration {
	return 10 * time.Minute
}

func (j *CleanupLogsJob) RetryCount() int {
	return 1
}

func (j *CleanupLogsJob) RetryDelay() time.Duration {
	return 5 * time.Minute
}
