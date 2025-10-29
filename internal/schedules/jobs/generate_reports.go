package jobs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"api-core/pkg/logger"
)

// GenerateReportsJob tạo reports định kỳ
type GenerateReportsJob struct{}

func (j *GenerateReportsJob) Name() string {
	return "generate-reports"
}

func (j *GenerateReportsJob) Run(ctx context.Context) error {
	jobLogger := logger.GetJobLogger(j.Name())
	jobLogger.Info().Msg("Starting generate reports job")

	// Tạo thư mục reports nếu chưa tồn tại
	reportsDir := "storages/reports"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		jobLogger.Error().Err(err).Msg("Failed to create reports directory")
		return err
	}

	// Tạo timestamp cho reports
	timestamp := time.Now().Format("20060102_150405")
	dateRange := time.Now().AddDate(0, 0, -7).Format("2006-01-02") + "_to_" + time.Now().Format("2006-01-02")

	// Danh sách các reports cần tạo
	reports := []struct {
		Name     string
		Filename string
		Type     string
	}{
		{"User Activity Report", fmt.Sprintf("user_activity_%s.csv", timestamp), "csv"},
		{"System Performance Report", fmt.Sprintf("system_performance_%s.xlsx", timestamp), "excel"},
		{"Error Log Summary", fmt.Sprintf("error_log_summary_%s.txt", timestamp), "text"},
		{"Weekly Statistics", fmt.Sprintf("weekly_stats_%s.json", timestamp), "json"},
	}

	generatedCount := 0

	for _, report := range reports {
		reportPath := filepath.Join(reportsDir, report.Filename)

		jobLogger.Info().
			Str("report_name", report.Name).
			Str("filename", report.Filename).
			Str("type", report.Type).
			Msg("Generating report")

		// Simulate report generation
		var content string
		switch report.Type {
		case "csv":
			content = fmt.Sprintf("Date,Users,Activity\n%s,150,2500\n%s,155,2600\n",
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
				time.Now().Format("2006-01-02"))
		case "excel":
			content = fmt.Sprintf("Weekly Performance Report\nGenerated: %s\nDate Range: %s\n\nUsers: 150\nActivity: 2500\n",
				time.Now().Format("2006-01-02 15:04:05"), dateRange)
		case "text":
			content = fmt.Sprintf("Error Log Summary\nGenerated: %s\n\nTotal Errors: 5\nCritical: 1\nWarning: 4\n",
				time.Now().Format("2006-01-02 15:04:05"))
		case "json":
			content = fmt.Sprintf(`{"report_type": "weekly_stats", "generated_at": "%s", "date_range": "%s", "total_users": 150, "total_activity": 2500}`,
				time.Now().Format(time.RFC3339), dateRange)
		}

		// Ghi file report
		if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
			jobLogger.Error().
				Err(err).
				Str("report_name", report.Name).
				Str("filename", report.Filename).
				Msg("Failed to generate report")
			continue
		}

		jobLogger.Info().
			Str("report_name", report.Name).
			Str("filename", report.Filename).
			Msg("Report generated successfully")

		generatedCount++

		// Simulate processing time
		time.Sleep(500 * time.Millisecond)
	}

	// Xóa reports cũ hơn 30 ngày
	cutoffTime := time.Now().AddDate(0, 0, -30)
	deletedCount := 0

	err := filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				jobLogger.Error().Err(err).Str("file", path).Msg("Failed to delete old report")
				return err
			}

			jobLogger.Info().Str("file", path).Msg("Deleted old report file")
			deletedCount++
		}

		return nil
	})

	if err != nil {
		jobLogger.Error().Err(err).Msg("Failed to cleanup old reports")
		return err
	}

	jobLogger.Info().
		Int("generated_count", generatedCount).
		Int("deleted_count", deletedCount).
		Int("total_reports", len(reports)).
		Msg("Generate reports job completed")

	return nil
}

func (j *GenerateReportsJob) Timeout() time.Duration {
	return 20 * time.Minute
}

func (j *GenerateReportsJob) RetryCount() int {
	return 1
}

func (j *GenerateReportsJob) RetryDelay() time.Duration {
	return 10 * time.Minute
}
