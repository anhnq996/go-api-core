package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Event represents a structured event to be sent to Loki
type Event struct {
	Action    string                 `json:"action"`    // create, update, delete, etc.
	Entity    string                 `json:"entity"`    // user, product, order, etc.
	EntityID  string                 `json:"entity_id"` // UUID of the entity
	UserID    string                 `json:"user_id"`   // ID of user performing action
	Data      map[string]interface{} `json:"data"`      // Additional data
	Timestamp time.Time              `json:"timestamp"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// Client handles sending events to Loki
type Client struct {
	lokiURL    string
	httpClient *http.Client
	labels     map[string]string
}

// Config for Loki client
type Config struct {
	URL         string            `json:"url"`
	Job         string            `json:"job"`
	Environment string            `json:"environment"`
	Labels      map[string]string `json:"labels"`
}

// NewClient creates a new Loki client
func NewClient(config Config) *Client {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	labels := map[string]string{
		"job":         config.Job,
		"environment": config.Environment,
		"host":        hostname,
	}

	// Merge additional labels
	for k, v := range config.Labels {
		labels[k] = v
	}

	return &Client{
		lokiURL: config.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		labels: labels,
	}
}

// SendEvent sends an event to Loki
func (c *Client) SendEvent(ctx context.Context, event Event) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Marshal event to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Prepare Loki push request
	timestamp := strconv.FormatInt(event.Timestamp.UnixNano(), 10)
	logLine := string(jsonData)

	pushReq := LokiPushRequest{
		Streams: []LokiStream{
			{
				Stream: c.labels,
				Values: [][]string{
					{timestamp, logLine},
				},
			},
		},
	}

	// Marshal to JSON
	requestData, err := json.Marshal(pushReq)
	if err != nil {
		return fmt.Errorf("failed to marshal push request: %w", err)
	}

	// Send to Loki
	url := c.lokiURL + "/loki/api/v1/push"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := json.Marshal(map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		})
		return fmt.Errorf("loki returned error: %s", string(body))
	}

	return nil
}

// SendEventAsync sends an event to Loki asynchronously
func (c *Client) SendEventAsync(ctx context.Context, event Event) {
	go func() {
		if err := c.SendEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Failed to send event to Loki: %v\n", err)
		}
	}()
}

// LokiPushRequest represents the Loki push API request
type LokiPushRequest struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream represents a log stream
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// Helper functions for common events

// CreateEvent creates a "create" event
func CreateEvent(entity, entityID, userID string, data map[string]interface{}) Event {
	return Event{
		Action:    "create",
		Entity:    entity,
		EntityID:  entityID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// UpdateEvent creates an "update" event
func UpdateEvent(entity, entityID, userID string, data map[string]interface{}) Event {
	return Event{
		Action:    "update",
		Entity:    entity,
		EntityID:  entityID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// DeleteEvent creates a "delete" event
func DeleteEvent(entity, entityID, userID string, data map[string]interface{}) Event {
	return Event{
		Action:    "delete",
		Entity:    entity,
		EntityID:  entityID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// LoginEvent creates a "login" event
func LoginEvent(userID, ip, userAgent string, data map[string]interface{}) Event {
	return Event{
		Action:    "login",
		Entity:    "user",
		EntityID:  userID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
	}
}

// LogoutEvent creates a "logout" event
func LogoutEvent(userID, ip, userAgent string, data map[string]interface{}) Event {
	return Event{
		Action:    "logout",
		Entity:    "user",
		EntityID:  userID,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
	}
}
