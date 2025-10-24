package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQQueue implements Queue using RabbitMQ
type RabbitMQQueue struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	name   string
	config *QueueConfig
}

// NewRabbitMQQueue creates a new RabbitMQ queue
func NewRabbitMQQueue(conn *amqp.Connection, name string, config *QueueConfig) (*RabbitMQQueue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQQueue{
		conn:   conn,
		ch:     ch,
		name:   name,
		config: config,
	}, nil
}

// Push adds a message to the queue
func (r *RabbitMQQueue) Push(ctx context.Context, message *Message) error {
	// Set message timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Prepare headers
	headers := make(amqp.Table)
	for k, v := range message.Headers {
		headers[k] = v
	}
	headers["retry_count"] = message.RetryCount
	headers["max_retries"] = message.MaxRetries
	headers["priority"] = message.Priority

	// Prepare publishing options
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         data,
		Headers:      headers,
		Timestamp:    message.Timestamp,
		DeliveryMode: amqp.Persistent, // Make message persistent
	}

	// Add delay if specified
	if message.Delay > 0 {
		// Use delayed message plugin or dead letter exchange
		publishing.Expiration = fmt.Sprintf("%d", message.Delay.Milliseconds())
	}

	// Publish message
	err = r.ch.Publish(
		"",     // exchange
		r.name, // routing key (queue name)
		false,  // mandatory
		false,  // immediate
		publishing,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Pop retrieves and removes a message from the queue
func (r *RabbitMQQueue) Pop(ctx context.Context) (*Message, error) {
	delivery, ok, err := r.ch.Get(r.name, false) // false = no auto-ack
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if !ok {
		return nil, nil // No messages
	}

	// Parse message
	var message Message
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		// Acknowledge the message even if parsing fails
		delivery.Ack(false)
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Extract headers
	if delivery.Headers != nil {
		if message.Headers == nil {
			message.Headers = make(map[string]string)
		}
		for k, v := range delivery.Headers {
			if str, ok := v.(string); ok {
				message.Headers[k] = str
			}
		}

		// Extract retry count and max retries
		if retryCount, ok := delivery.Headers["retry_count"].(int); ok {
			message.RetryCount = retryCount
		}
		if maxRetries, ok := delivery.Headers["max_retries"].(int); ok {
			message.MaxRetries = maxRetries
		}
		if priority, ok := delivery.Headers["priority"].(int); ok {
			message.Priority = priority
		}
	}

	// Acknowledge the message
	if err := delivery.Ack(false); err != nil {
		return nil, fmt.Errorf("failed to acknowledge message: %w", err)
	}

	return &message, nil
}

// PopWithTimeout retrieves a message with timeout
func (r *RabbitMQQueue) PopWithTimeout(ctx context.Context, timeout time.Duration) (*Message, error) {
	// RabbitMQ doesn't have built-in timeout for Get, so we use a goroutine
	done := make(chan *Message, 1)
	errChan := make(chan error, 1)

	go func() {
		message, err := r.Pop(ctx)
		if err != nil {
			errChan <- err
			return
		}
		done <- message
	}()

	select {
	case message := <-done:
		return message, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, nil // Timeout
	}
}

// Peek retrieves a message without removing it
func (r *RabbitMQQueue) Peek(ctx context.Context) (*Message, error) {
	delivery, ok, err := r.ch.Get(r.name, true) // true = auto-ack (peek)
	if err != nil {
		return nil, fmt.Errorf("failed to peek message: %w", err)
	}

	if !ok {
		return nil, nil // No messages
	}

	// Parse message
	var message Message
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Extract headers
	if delivery.Headers != nil {
		if message.Headers == nil {
			message.Headers = make(map[string]string)
		}
		for k, v := range delivery.Headers {
			if str, ok := v.(string); ok {
				message.Headers[k] = str
			}
		}
	}

	return &message, nil
}

// Size returns the number of messages in the queue
func (r *RabbitMQQueue) Size(ctx context.Context) (int64, error) {
	queue, err := r.ch.QueueInspect(r.name)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return int64(queue.Messages), nil
}

// Clear removes all messages from the queue
func (r *RabbitMQQueue) Clear(ctx context.Context) error {
	_, err := r.ch.QueuePurge(r.name, false)
	if err != nil {
		return fmt.Errorf("failed to purge queue: %w", err)
	}

	return nil
}

// Close closes the queue connection
func (r *RabbitMQQueue) Close() error {
	if r.ch != nil {
		return r.ch.Close()
	}
	return nil
}

// GetName returns the queue name
func (r *RabbitMQQueue) GetName() string {
	return r.name
}

// RabbitMQQueueManager implements QueueManager using RabbitMQ
type RabbitMQQueueManager struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	config *QueueConfig
	queues map[string]Queue
}

// NewRabbitMQQueueManager creates a new RabbitMQ queue manager
func NewRabbitMQQueueManager(conn *amqp.Connection, config *QueueConfig) (*RabbitMQQueueManager, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQQueueManager{
		conn:   conn,
		ch:     ch,
		config: config,
		queues: make(map[string]Queue),
	}, nil
}

// CreateQueue creates a new queue
func (r *RabbitMQQueueManager) CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error) {
	if options == nil {
		options = r.config.DefaultQueueOptions
	}

	// Set default options
	if options == nil {
		options = &QueueOptions{
			Durable: true,
		}
	}

	// Declare queue
	_, err := r.ch.QueueDeclare(
		name,                                  // name
		options.Durable,                       // durable
		options.AutoDelete,                    // delete when unused
		options.Exclusive,                     // exclusive
		options.NoWait,                        // no-wait
		convertToAMQPTable(options.Arguments), // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Create queue instance
	queue, err := NewRabbitMQQueue(r.conn, name, r.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue instance: %w", err)
	}

	r.queues[name] = queue
	return queue, nil
}

// GetQueue returns an existing queue
func (r *RabbitMQQueueManager) GetQueue(name string) (Queue, error) {
	if queue, exists := r.queues[name]; exists {
		return queue, nil
	}

	// Create queue if it doesn't exist
	return r.CreateQueue(context.Background(), name, nil)
}

// DeleteQueue deletes a queue
func (r *RabbitMQQueueManager) DeleteQueue(ctx context.Context, name string) error {
	// Delete queue from RabbitMQ
	_, err := r.ch.QueueDelete(name, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	delete(r.queues, name)
	return nil
}

// ListQueues returns a list of all queues
func (r *RabbitMQQueueManager) ListQueues(ctx context.Context) ([]string, error) {
	// RabbitMQ doesn't have a direct way to list queues via AMQP
	// This would typically require management API or admin interface
	// For now, return the queues we know about
	queues := make([]string, 0, len(r.queues))
	for name := range r.queues {
		queues = append(queues, name)
	}

	return queues, nil
}

// Close closes all queue connections
func (r *RabbitMQQueueManager) Close() error {
	if r.ch != nil {
		if err := r.ch.Close(); err != nil {
			return err
		}
	}
	return nil
}

// convertToAMQPTable converts map[string]interface{} to amqp.Table
func convertToAMQPTable(args map[string]interface{}) amqp.Table {
	if args == nil {
		return nil
	}

	table := make(amqp.Table)
	for k, v := range args {
		table[k] = v
	}

	return table
}
