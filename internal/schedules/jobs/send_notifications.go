package jobs

import (
	"context"
	"time"

	"anhnq/api-core/pkg/logger"
)

// SendNotificationsJob gửi notifications
type SendNotificationsJob struct{}

func (j *SendNotificationsJob) Name() string {
	return "send-notifications"
}

func (j *SendNotificationsJob) Run(ctx context.Context) error {
	jobLogger := logger.GetJobLogger(j.Name())
	jobLogger.Info().Msg("Starting send notifications job")

	// Simulate checking for pending notifications
	// Trong thực tế, bạn sẽ query database để lấy notifications cần gửi
	pendingNotifications := []struct {
		ID      string
		UserID  string
		Message string
		Type    string
	}{
		{"1", "user1", "Welcome to our service!", "welcome"},
		{"2", "user2", "Your order has been shipped!", "shipping"},
		{"3", "user3", "Payment reminder", "payment"},
	}

	// Simulate sending notifications
	for _, notification := range pendingNotifications {
		jobLogger.Info().
			Str("notification_id", notification.ID).
			Str("user_id", notification.UserID).
			Str("type", notification.Type).
			Str("message", notification.Message).
			Msg("Sending notification")

		// Simulate sending delay
		time.Sleep(100 * time.Millisecond)

		// Simulate success/failure
		if notification.ID == "2" {
			jobLogger.Error().
				Str("notification_id", notification.ID).
				Msg("Failed to send notification")
			continue
		}

		jobLogger.Info().
			Str("notification_id", notification.ID).
			Msg("Notification sent successfully")
	}

	jobLogger.Info().Int("total_notifications", len(pendingNotifications)).Msg("Send notifications job completed")
	return nil
}

func (j *SendNotificationsJob) Timeout() time.Duration {
	return 5 * time.Minute
}

func (j *SendNotificationsJob) RetryCount() int {
	return 3
}

func (j *SendNotificationsJob) RetryDelay() time.Duration {
	return 1 * time.Minute
}
