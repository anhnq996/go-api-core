package jobs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"api-core/pkg/logger"
)

// HealthCheckJob kiểm tra health của các services
type HealthCheckJob struct{}

func (j *HealthCheckJob) Name() string {
	return "health-check"
}

func (j *HealthCheckJob) Run(ctx context.Context) error {
	jobLogger := logger.GetJobLogger(j.Name())
	jobLogger.Info().Msg("Starting health check job")

	// Danh sách các services cần kiểm tra
	services := []struct {
		Name string
		URL  string
	}{
		{"Database", "http://localhost:3000/api/v1/health/database"},
		{"Redis", "http://localhost:3000/api/v1/health/redis"},
		{"API Server", "http://localhost:3000/api/v1/health"},
	}

	healthyCount := 0
	unhealthyCount := 0

	for _, service := range services {
		start := time.Now()

		// Tạo HTTP client với timeout
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// Gửi request kiểm tra health
		resp, err := client.Get(service.URL)
		duration := time.Since(start)

		if err != nil {
			jobLogger.Error().
				Str("service", service.Name).
				Str("url", service.URL).
				Err(err).
				Dur("duration", duration).
				Msg("Service health check failed")
			unhealthyCount++
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			jobLogger.Info().
				Str("service", service.Name).
				Str("url", service.URL).
				Int("status_code", resp.StatusCode).
				Dur("duration", duration).
				Msg("Service is healthy")
			healthyCount++
		} else {
			jobLogger.Error().
				Str("service", service.Name).
				Str("url", service.URL).
				Int("status_code", resp.StatusCode).
				Dur("duration", duration).
				Msg("Service is unhealthy")
			unhealthyCount++
		}
	}

	// Log tổng kết
	jobLogger.Info().
		Int("healthy_count", healthyCount).
		Int("unhealthy_count", unhealthyCount).
		Int("total_services", len(services)).
		Msg("Health check job completed")

	// Nếu có service nào unhealthy, return error để trigger retry
	if unhealthyCount > 0 {
		return fmt.Errorf("health check failed for %d services", unhealthyCount)
	}

	return nil
}

func (j *HealthCheckJob) Timeout() time.Duration {
	return 5 * time.Minute
}

func (j *HealthCheckJob) RetryCount() int {
	return 3
}

func (j *HealthCheckJob) RetryDelay() time.Duration {
	return 2 * time.Minute
}
