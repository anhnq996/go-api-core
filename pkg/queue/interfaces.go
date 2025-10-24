package queue

import (
	"context"
	"time"
)

// Message represents a message in the queue
type Message struct {
	ID         string            `json:"id"`
	Data       []byte            `json:"data"`
	Headers    map[string]string `json:"headers,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
	RetryCount int               `json:"retry_count"`
	MaxRetries int               `json:"max_retries"`
	Delay      time.Duration     `json:"delay,omitempty"`
	Priority   int               `json:"priority,omitempty"`
}

// Job represents a job to be processed
type Job interface {
	// GetID returns the unique identifier for the job
	GetID() string

	// GetData returns the job data
	GetData() []byte

	// GetHeaders returns the job headers
	GetHeaders() map[string]string

	// GetMaxRetries returns the maximum number of retries
	GetMaxRetries() int

	// GetDelay returns the delay before processing
	GetDelay() time.Duration

	// GetPriority returns the job priority
	GetPriority() int

	// Process processes the job
	Process(ctx context.Context) error
}

// Queue represents a message queue
type Queue interface {
	// Push adds a message to the queue
	Push(ctx context.Context, message *Message) error

	// Pop retrieves and removes a message from the queue
	Pop(ctx context.Context) (*Message, error)

	// PopWithTimeout retrieves a message with timeout
	PopWithTimeout(ctx context.Context, timeout time.Duration) (*Message, error)

	// Peek retrieves a message without removing it
	Peek(ctx context.Context) (*Message, error)

	// Size returns the number of messages in the queue
	Size(ctx context.Context) (int64, error)

	// Clear removes all messages from the queue
	Clear(ctx context.Context) error

	// Close closes the queue connection
	Close() error

	// GetName returns the queue name
	GetName() string
}

// Consumer represents a message consumer
type Consumer interface {
	// Start starts consuming messages from the queue
	Start(ctx context.Context) error

	// Stop stops consuming messages
	Stop() error

	// IsRunning returns true if the consumer is running
	IsRunning() bool

	// GetQueue returns the queue being consumed
	GetQueue() Queue

	// GetHandler returns the message handler
	GetHandler() MessageHandler
}

// MessageHandler processes messages from the queue
type MessageHandler interface {
	// Handle processes a message
	Handle(ctx context.Context, message *Message) error

	// OnError handles errors during message processing
	OnError(ctx context.Context, message *Message, err error) error
}

// Producer represents a message producer
type Producer interface {
	// Publish publishes a message to the queue
	Publish(ctx context.Context, message *Message) error

	// PublishBatch publishes multiple messages to the queue
	PublishBatch(ctx context.Context, messages []*Message) error

	// Close closes the producer connection
	Close() error

	// GetQueue returns the queue being published to
	GetQueue() Queue
}

// QueueManager manages multiple queues
type QueueManager interface {
	// CreateQueue creates a new queue
	CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error)

	// GetQueue returns an existing queue
	GetQueue(name string) (Queue, error)

	// DeleteQueue deletes a queue
	DeleteQueue(ctx context.Context, name string) error

	// ListQueues returns a list of all queues
	ListQueues(ctx context.Context) ([]string, error)

	// Close closes all queue connections
	Close() error
}

// QueueOptions represents options for creating a queue
type QueueOptions struct {
	// Durable specifies if the queue should survive broker restarts
	Durable bool `json:"durable"`

	// AutoDelete specifies if the queue should be deleted when unused
	AutoDelete bool `json:"auto_delete"`

	// Exclusive specifies if the queue should be exclusive to the connection
	Exclusive bool `json:"exclusive"`

	// NoWait specifies if the queue creation should not wait for confirmation
	NoWait bool `json:"no_wait"`

	// Arguments specifies additional queue arguments
	Arguments map[string]interface{} `json:"arguments,omitempty"`

	// MaxLength specifies the maximum number of messages in the queue
	MaxLength int64 `json:"max_length,omitempty"`

	// MessageTTL specifies the TTL for messages in the queue
	MessageTTL time.Duration `json:"message_ttl,omitempty"`

	// DeadLetterExchange specifies the dead letter exchange
	DeadLetterExchange string `json:"dead_letter_exchange,omitempty"`

	// DeadLetterRoutingKey specifies the dead letter routing key
	DeadLetterRoutingKey string `json:"dead_letter_routing_key,omitempty"`
}

// ConsumerOptions represents options for creating a consumer
type ConsumerOptions struct {
	// AutoAck specifies if messages should be automatically acknowledged
	AutoAck bool `json:"auto_ack"`

	// Exclusive specifies if the consumer should be exclusive
	Exclusive bool `json:"exclusive"`

	// NoLocal specifies if the consumer should not receive messages published by the same connection
	NoLocal bool `json:"no_local"`

	// NoWait specifies if the consumer should not wait for confirmation
	NoWait bool `json:"no_wait"`

	// PrefetchCount specifies the number of messages to prefetch
	PrefetchCount int `json:"prefetch_count"`

	// PrefetchSize specifies the prefetch size in bytes
	PrefetchSize int `json:"prefetch_size"`

	// Global specifies if the prefetch should be global
	Global bool `json:"global"`

	// Concurrency specifies the number of concurrent workers
	Concurrency int `json:"concurrency"`

	// RetryDelay specifies the delay between retries
	RetryDelay time.Duration `json:"retry_delay"`

	// MaxRetries specifies the maximum number of retries
	MaxRetries int `json:"max_retries"`
}

// QueueBackend represents a queue backend implementation
type QueueBackend interface {
	// Connect connects to the queue backend
	Connect(ctx context.Context) error

	// Disconnect disconnects from the queue backend
	Disconnect() error

	// IsConnected returns true if connected
	IsConnected() bool

	// CreateQueue creates a queue
	CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error)

	// DeleteQueue deletes a queue
	DeleteQueue(ctx context.Context, name string) error

	// ListQueues lists all queues
	ListQueues(ctx context.Context) ([]string, error)
}

// QueueType represents the type of queue backend
type QueueType string

const (
	QueueTypeRedis    QueueType = "redis"
	QueueTypeRabbitMQ QueueType = "rabbitmq"
	QueueTypeMemory   QueueType = "memory"
)

// QueueConfig represents the configuration for a queue backend
type QueueConfig struct {
	Type     QueueType `json:"type"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Username string    `json:"username,omitempty"`
	Password string    `json:"password,omitempty"`
	Database int       `json:"database,omitempty"`
	VHost    string    `json:"vhost,omitempty"`

	// Connection options
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
	MaxIdleConns   int           `json:"max_idle_conns"`
	MaxActiveConns int           `json:"max_active_conns"`

	// Queue options
	DefaultQueueOptions *QueueOptions `json:"default_queue_options,omitempty"`

	// Consumer options
	DefaultConsumerOptions *ConsumerOptions `json:"default_consumer_options,omitempty"`
}
