package jobs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"anhnq/api-core/pkg/logger"
)

// CleanupTempFilesJob xóa temp files
type CleanupTempFilesJob struct{}

func (j *CleanupTempFilesJob) Name() string {
	return "cleanup-temp-files"
}

func (j *CleanupTempFilesJob) Run(ctx context.Context) error {
	jobLogger := logger.GetJobLogger(j.Name())
	jobLogger.Info().Msg("Starting cleanup temp files job")

	// Các thư mục temp cần cleanup
	tempDirs := []string{
		"storages/temp",
		"storages/uploads/temp",
		"tmp",
	}

	deletedCount := 0
	totalSize := int64(0)

	for _, tempDir := range tempDirs {
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Chỉ xử lý files
			if info.IsDir() {
				return nil
			}

			// Xóa files cũ hơn 1 giờ
			if time.Since(info.ModTime()) > time.Hour {
				// Kiểm tra extension để đảm bảo là temp file
				ext := strings.ToLower(filepath.Ext(path))
				tempExtensions := []string{".tmp", ".temp", ".cache", ".log"}

				isTempFile := false
				for _, tempExt := range tempExtensions {
					if ext == tempExt {
						isTempFile = true
						break
					}
				}

				// Hoặc files có tên chứa "temp"
				if strings.Contains(strings.ToLower(filepath.Base(path)), "temp") {
					isTempFile = true
				}

				if isTempFile {
					if err := os.Remove(path); err != nil {
						jobLogger.Error().Err(err).Str("file", path).Msg("Failed to delete temp file")
						return err
					}

					jobLogger.Info().
						Str("file", path).
						Int64("size", info.Size()).
						Msg("Deleted temp file")

					deletedCount++
					totalSize += info.Size()
				}
			}

			return nil
		})

		if err != nil {
			jobLogger.Error().Err(err).Str("dir", tempDir).Msg("Failed to cleanup temp directory")
			return err
		}
	}

	jobLogger.Info().
		Int("deleted_count", deletedCount).
		Int64("total_size_mb", totalSize/1024/1024).
		Msg("Cleanup temp files job completed")

	return nil
}

func (j *CleanupTempFilesJob) Timeout() time.Duration {
	return 15 * time.Minute
}

func (j *CleanupTempFilesJob) RetryCount() int {
	return 1
}

func (j *CleanupTempFilesJob) RetryDelay() time.Duration {
	return 5 * time.Minute
}
