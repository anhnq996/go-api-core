package loki

import (
	"context"
)

// EventLogger interface for logging events
type EventLogger interface {
	LogEvent(ctx context.Context, event Event) error
	LogEventAsync(ctx context.Context, event Event)

	// Convenience methods
	LogCreate(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error
	LogUpdate(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error
	LogDelete(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error
	LogLogin(ctx context.Context, userID, ip, userAgent string, data map[string]interface{}) error
	LogLogout(ctx context.Context, userID, ip, userAgent string, data map[string]interface{}) error
}

// Service provides event logging functionality
type Service struct {
	client *Client
}

// NewService creates a new event logging service
func NewService(config Config) *Service {
	return &Service{
		client: NewClient(config),
	}
}

// LogEvent logs an event to Loki
func (s *Service) LogEvent(ctx context.Context, event Event) error {
	return s.client.SendEvent(ctx, event)
}

// LogEventAsync logs an event to Loki asynchronously
func (s *Service) LogEventAsync(ctx context.Context, event Event) {
	s.client.SendEventAsync(ctx, event)
}

// LogCreate logs a create event
func (s *Service) LogCreate(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error {
	event := CreateEvent(entity, entityID, userID, data)
	return s.LogEvent(ctx, event)
}

// LogUpdate logs an update event
func (s *Service) LogUpdate(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error {
	event := UpdateEvent(entity, entityID, userID, data)
	return s.LogEvent(ctx, event)
}

// LogDelete logs a delete event
func (s *Service) LogDelete(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error {
	event := DeleteEvent(entity, entityID, userID, data)
	return s.LogEvent(ctx, event)
}

// LogLogin logs a login event
func (s *Service) LogLogin(ctx context.Context, userID, ip, userAgent string, data map[string]interface{}) error {
	event := LoginEvent(userID, ip, userAgent, data)
	return s.LogEvent(ctx, event)
}

// LogLogout logs a logout event
func (s *Service) LogLogout(ctx context.Context, userID, ip, userAgent string, data map[string]interface{}) error {
	event := LogoutEvent(userID, ip, userAgent, data)
	return s.LogEvent(ctx, event)
}

// Global service instance
var GlobalService EventLogger

// Init initializes the global Loki service
func Init(config Config) {
	GlobalService = NewService(config)
}

// Helper functions using global service

// LogEvent logs an event using global service
func LogEvent(ctx context.Context, event Event) error {
	if GlobalService == nil {
		return nil // Silently ignore if not initialized
	}
	return GlobalService.LogEvent(ctx, event)
}

// LogEventAsync logs an event asynchronously using global service
func LogEventAsync(ctx context.Context, event Event) {
	if GlobalService == nil {
		return // Silently ignore if not initialized
	}
	GlobalService.LogEventAsync(ctx, event)
}

// LogCreate logs a create event using global service
func LogCreate(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogCreate(ctx, entity, entityID, userID, data)
}

// LogUpdate logs an update event using global service
func LogUpdate(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogUpdate(ctx, entity, entityID, userID, data)
}

// LogDelete logs a delete event using global service
func LogDelete(ctx context.Context, entity, entityID, userID string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogDelete(ctx, entity, entityID, userID, data)
}

// LogLogin logs a login event using global service
func LogLogin(ctx context.Context, userID, ip, userAgent string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogLogin(ctx, userID, ip, userAgent, data)
}

// LogLogout logs a logout event using global service
func LogLogout(ctx context.Context, userID, ip, userAgent string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogLogout(ctx, userID, ip, userAgent, data)
}
