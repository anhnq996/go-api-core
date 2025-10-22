package fcm

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Client là FCM client để gửi notifications
type Client struct {
	app             *firebase.App
	messagingClient *messaging.Client
	config          *Config
}

// Config cấu hình cho FCM client
type Config struct {
	CredentialsFile string        // Đường dẫn tới file credentials JSON của Firebase
	Timeout         time.Duration // Timeout cho mỗi request
	ProjectID       string        // Firebase project ID (optional, có thể lấy từ credentials)
}

// NewClient tạo FCM client mới
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config không được để trống")
	}

	if cfg.CredentialsFile == "" {
		return nil, fmt.Errorf("credentials file không được để trống")
	}

	// Set default timeout nếu không có
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	// Khởi tạo Firebase app
	opt := option.WithCredentialsFile(cfg.CredentialsFile)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo Firebase app: %w", err)
	}

	// Khởi tạo messaging client
	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo messaging client: %w", err)
	}

	return &Client{
		app:             app,
		messagingClient: messagingClient,
		config:          cfg,
	}, nil
}

// SendToToken gửi notification đến một device token cụ thể
func (c *Client) SendToToken(ctx context.Context, token string, notification *Notification, data map[string]string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token không được để trống")
	}

	message := &messaging.Message{
		Token: token,
		Data:  data,
	}

	if notification != nil {
		message.Notification = &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		}
	}

	// Áp dụng platform specific config nếu có
	if notification != nil && notification.Android != nil {
		message.Android = notification.Android
	}

	if notification != nil && notification.APNS != nil {
		message.APNS = notification.APNS
	}

	if notification != nil && notification.Webpush != nil {
		message.Webpush = notification.Webpush
	}

	// Gửi message
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	messageID, err := c.messagingClient.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("không thể gửi message: %w", err)
	}

	return messageID, nil
}

// SendToTokens gửi notification đến nhiều device tokens (tối đa 500 tokens)
func (c *Client) SendToTokens(ctx context.Context, tokens []string, notification *Notification, data map[string]string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("danh sách tokens không được để trống")
	}

	if len(tokens) > 500 {
		return nil, fmt.Errorf("chỉ được gửi tối đa 500 tokens mỗi lần")
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Data:   data,
	}

	if notification != nil {
		message.Notification = &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		}
	}

	// Áp dụng platform specific config nếu có
	if notification != nil && notification.Android != nil {
		message.Android = notification.Android
	}

	if notification != nil && notification.APNS != nil {
		message.APNS = notification.APNS
	}

	if notification != nil && notification.Webpush != nil {
		message.Webpush = notification.Webpush
	}

	// Gửi multicast message
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	response, err := c.messagingClient.SendEachForMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("không thể gửi multicast message: %w", err)
	}

	return response, nil
}

// SendToTopic gửi notification đến một topic
func (c *Client) SendToTopic(ctx context.Context, topic string, notification *Notification, data map[string]string) (string, error) {
	if topic == "" {
		return "", fmt.Errorf("topic không được để trống")
	}

	message := &messaging.Message{
		Topic: topic,
		Data:  data,
	}

	if notification != nil {
		message.Notification = &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		}
	}

	// Áp dụng platform specific config nếu có
	if notification != nil && notification.Android != nil {
		message.Android = notification.Android
	}

	if notification != nil && notification.APNS != nil {
		message.APNS = notification.APNS
	}

	if notification != nil && notification.Webpush != nil {
		message.Webpush = notification.Webpush
	}

	// Gửi message
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	messageID, err := c.messagingClient.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("không thể gửi message đến topic: %w", err)
	}

	return messageID, nil
}

// SendToCondition gửi notification dựa trên điều kiện topic
func (c *Client) SendToCondition(ctx context.Context, condition string, notification *Notification, data map[string]string) (string, error) {
	if condition == "" {
		return "", fmt.Errorf("condition không được để trống")
	}

	message := &messaging.Message{
		Condition: condition,
		Data:      data,
	}

	if notification != nil {
		message.Notification = &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		}
	}

	// Áp dụng platform specific config nếu có
	if notification != nil && notification.Android != nil {
		message.Android = notification.Android
	}

	if notification != nil && notification.APNS != nil {
		message.APNS = notification.APNS
	}

	if notification != nil && notification.Webpush != nil {
		message.Webpush = notification.Webpush
	}

	// Gửi message
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	messageID, err := c.messagingClient.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("không thể gửi message với condition: %w", err)
	}

	return messageID, nil
}

// SubscribeToTopic đăng ký tokens vào một topic
func (c *Client) SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("danh sách tokens không được để trống")
	}

	if topic == "" {
		return nil, fmt.Errorf("topic không được để trống")
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	response, err := c.messagingClient.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return nil, fmt.Errorf("không thể subscribe tokens vào topic: %w", err)
	}

	return response, nil
}

// UnsubscribeFromTopic hủy đăng ký tokens khỏi một topic
func (c *Client) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("danh sách tokens không được để trống")
	}

	if topic == "" {
		return nil, fmt.Errorf("topic không được để trống")
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	response, err := c.messagingClient.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return nil, fmt.Errorf("không thể unsubscribe tokens khỏi topic: %w", err)
	}

	return response, nil
}

// SendAll gửi nhiều messages khác nhau (tối đa 500 messages)
func (c *Client) SendAll(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("danh sách messages không được để trống")
	}

	if len(messages) > 500 {
		return nil, fmt.Errorf("chỉ được gửi tối đa 500 messages mỗi lần")
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	response, err := c.messagingClient.SendEach(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("không thể gửi batch messages: %w", err)
	}

	return response, nil
}

// SendDryRun gửi message ở chế độ dry-run để test mà không gửi thực sự
func (c *Client) SendDryRun(ctx context.Context, token string, notification *Notification, data map[string]string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token không được để trống")
	}

	message := &messaging.Message{
		Token: token,
		Data:  data,
	}

	if notification != nil {
		message.Notification = &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	messageID, err := c.messagingClient.SendDryRun(ctx, message)
	if err != nil {
		return "", fmt.Errorf("không thể gửi dry-run message: %w", err)
	}

	return messageID, nil
}
