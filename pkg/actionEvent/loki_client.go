package actionEvent

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

// LokiClientImpl implements LokiClient interface
type LokiClientImpl struct {
	lokiURL    string
	httpClient *http.Client
	labels     map[string]string
}

// NewLokiClient creates a new Loki client
func NewLokiClient(lokiURL string, defaultLabels map[string]string) *LokiClientImpl {
	return &LokiClientImpl{
		lokiURL: lokiURL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second, // Shorter timeout for async
		},
		labels: defaultLabels,
	}
}

// PushEventAsync pushes an event to Loki asynchronously
func (c *LokiClientImpl) PushEventAsync(ctx context.Context, job string, event Event) error {
	go func() {
		if err := c.pushEventSync(ctx, job, event); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Failed to push event to Loki: %v\n", err)
		}
	}()
	return nil // Always return nil for async operation
}

// pushEventSync pushes an event to Loki synchronously (internal use)
func (c *LokiClientImpl) pushEventSync(ctx context.Context, job string, event Event) error {
	// Marshal event to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Prepare labels with dynamic job
	labels := make(map[string]string)
	for k, v := range c.labels {
		labels[k] = v
	}
	labels["job"] = job

	// Prepare Loki push request
	timestamp := strconv.FormatInt(event.Timestamp.UnixNano(), 10)
	logLine := string(jsonData)

	pushReq := LokiPushRequest{
		Streams: []LokiStream{
			{
				Stream: labels,
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
		return fmt.Errorf("loki returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
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
