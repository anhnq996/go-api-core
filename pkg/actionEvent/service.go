package actionEvent

import (
	"context"
	"time"
)

// EventData represents the data structure for action events
type EventData struct {
	Old map[string]interface{} `json:"old,omitempty"` // Previous data (null for create)
	New map[string]interface{} `json:"new,omitempty"` // New data (same as old for delete)
}

// Event represents a structured action event
type Event struct {
	Action    string    `json:"action"`    // create, update, delete, login, logout, etc.
	Entity    string    `json:"entity"`    // user, product, order, etc.
	EntityID  string    `json:"entity_id"` // UUID of the entity
	UserID    string    `json:"user_id"`   // ID of user performing action
	Data      EventData `json:"data"`      // Old and new data
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Job       string    `json:"job"` // Dynamic job name
}

// EventLogger interface for logging action events
type EventLogger interface {
	LogEvent(ctx context.Context, event Event) error
	LogEventAsync(ctx context.Context, event Event)

	// Convenience methods with dynamic job
	LogCreate(ctx context.Context, job, entity, entityID, userID string, newData map[string]interface{}) error
	LogUpdate(ctx context.Context, job, entity, entityID, userID string, oldData, newData map[string]interface{}) error
	LogDelete(ctx context.Context, job, entity, entityID, userID string, data map[string]interface{}) error
	LogLogin(ctx context.Context, job, userID, ip, userAgent string, data map[string]interface{}) error
	LogLogout(ctx context.Context, job, userID, ip, userAgent string, data map[string]interface{}) error
}

// Service provides action event logging functionality
type Service struct {
	lokiClient LokiClient
}

// LokiClient interface for pushing events to Loki
type LokiClient interface {
	PushEventAsync(ctx context.Context, job string, event Event) error
}

// NewService creates a new action event service
func NewService(lokiClient LokiClient) *Service {
	return &Service{
		lokiClient: lokiClient,
	}
}

// LogEvent logs an event to Loki
func (s *Service) LogEvent(ctx context.Context, event Event) error {
	return s.lokiClient.PushEventAsync(ctx, event.Job, event)
}

// LogEventAsync logs an event to Loki asynchronously
func (s *Service) LogEventAsync(ctx context.Context, event Event) {
	go func() {
		if err := s.LogEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			// You can use your logger here if needed
		}
	}()
}

// LogCreate logs a create event
func (s *Service) LogCreate(ctx context.Context, job, entity, entityID, userID string, newData map[string]interface{}) error {
	event := Event{
		Action:   "create",
		Entity:   entity,
		EntityID: entityID,
		UserID:   userID,
		Data: EventData{
			Old: nil,     // No old data for create
			New: newData, // New data
		},
		Timestamp: time.Now(),
		Job:       job,
	}
	return s.LogEvent(ctx, event)
}

// LogUpdate logs an update event
func (s *Service) LogUpdate(ctx context.Context, job, entity, entityID, userID string, oldData, newData map[string]interface{}) error {
	event := Event{
		Action:   "update",
		Entity:   entity,
		EntityID: entityID,
		UserID:   userID,
		Data: EventData{
			Old: oldData, // Previous data
			New: newData, // Updated data
		},
		Timestamp: time.Now(),
		Job:       job,
	}
	return s.LogEvent(ctx, event)
}

// LogDelete logs a delete event
func (s *Service) LogDelete(ctx context.Context, job, entity, entityID, userID string, data map[string]interface{}) error {
	event := Event{
		Action:   "delete",
		Entity:   entity,
		EntityID: entityID,
		UserID:   userID,
		Data: EventData{
			Old: data, // Old data (same as new for delete)
			New: data, // Same data (for delete)
		},
		Timestamp: time.Now(),
		Job:       job,
	}
	return s.LogEvent(ctx, event)
}

// LogLogin logs a login event
func (s *Service) LogLogin(ctx context.Context, job, userID, ip, userAgent string, data map[string]interface{}) error {
	event := Event{
		Action:   "login",
		Entity:   "user",
		EntityID: userID,
		UserID:   userID,
		Data: EventData{
			New: data, // Login data
		},
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		Job:       job,
	}
	return s.LogEvent(ctx, event)
}

// LogLogout logs a logout event
func (s *Service) LogLogout(ctx context.Context, job, userID, ip, userAgent string, data map[string]interface{}) error {
	event := Event{
		Action:   "logout",
		Entity:   "user",
		EntityID: userID,
		UserID:   userID,
		Data: EventData{
			New: data, // Logout data
		},
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		Job:       job,
	}
	return s.LogEvent(ctx, event)
}

// Global service instance
var GlobalService EventLogger

// Init initializes the global action event service
func Init(lokiClient LokiClient) {
	GlobalService = NewService(lokiClient)
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
func LogCreate(ctx context.Context, job, entity, entityID, userID string, newData map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogCreate(ctx, job, entity, entityID, userID, newData)
}

// LogUpdate logs an update event using global service
func LogUpdate(ctx context.Context, job, entity, entityID, userID string, oldData, newData map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogUpdate(ctx, job, entity, entityID, userID, oldData, newData)
}

// LogDelete logs a delete event using global service
func LogDelete(ctx context.Context, job, entity, entityID, userID string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogDelete(ctx, job, entity, entityID, userID, data)
}

// LogLogin logs a login event using global service
func LogLogin(ctx context.Context, job, userID, ip, userAgent string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogLogin(ctx, job, userID, ip, userAgent, data)
}

// LogLogout logs a logout event using global service
func LogLogout(ctx context.Context, job, userID, ip, userAgent string, data map[string]interface{}) error {
	if GlobalService == nil {
		return nil
	}
	return GlobalService.LogLogout(ctx, job, userID, ip, userAgent, data)
}
