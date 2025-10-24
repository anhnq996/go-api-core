package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
)

// QueueManagerImpl implements QueueManager
type QueueManagerImpl struct {
	backend QueueBackend
	queues  map[string]Queue
	config  *QueueConfig
}

// NewQueueManager creates a new queue manager
func NewQueueManager(config *QueueConfig) (*QueueManagerImpl, error) {
	var backend QueueBackend
	var err error

	switch config.Type {
	case QueueTypeRedis:
		backend, err = NewRedisBackend(config)
	case QueueTypeRabbitMQ:
		backend, err = NewRabbitMQBackend(config)
	case QueueTypeMemory:
		backend, err = NewMemoryBackend(config)
	default:
		return nil, fmt.Errorf("unsupported queue type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	return &QueueManagerImpl{
		backend: backend,
		queues:  make(map[string]Queue),
		config:  config,
	}, nil
}

// CreateQueue creates a new queue
func (q *QueueManagerImpl) CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error) {
	queue, err := q.backend.CreateQueue(ctx, name, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	q.queues[name] = queue
	return queue, nil
}

// GetQueue returns an existing queue
func (q *QueueManagerImpl) GetQueue(name string) (Queue, error) {
	if queue, exists := q.queues[name]; exists {
		return queue, nil
	}

	// Create queue if it doesn't exist
	return q.CreateQueue(context.Background(), name, q.config.DefaultQueueOptions)
}

// DeleteQueue deletes a queue
func (q *QueueManagerImpl) DeleteQueue(ctx context.Context, name string) error {
	if err := q.backend.DeleteQueue(ctx, name); err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	delete(q.queues, name)
	return nil
}

// ListQueues returns a list of all queues
func (q *QueueManagerImpl) ListQueues(ctx context.Context) ([]string, error) {
	return q.backend.ListQueues(ctx)
}

// Close closes all queue connections
func (q *QueueManagerImpl) Close() error {
	return q.backend.Disconnect()
}

// RedisBackend implements QueueBackend using Redis
type RedisBackend struct {
	client *redis.Client
	config *QueueConfig
}

// NewRedisBackend creates a new Redis backend
func NewRedisBackend(config *QueueConfig) (*RedisBackend, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.ConnectTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
		PoolSize:     config.MaxActiveConns,
	})

	return &RedisBackend{
		client: client,
		config: config,
	}, nil
}

// Connect connects to Redis
func (r *RedisBackend) Connect(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Disconnect disconnects from Redis
func (r *RedisBackend) Disconnect() error {
	return r.client.Close()
}

// IsConnected returns true if connected
func (r *RedisBackend) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return r.client.Ping(ctx).Err() == nil
}

// CreateQueue creates a queue
func (r *RedisBackend) CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error) {
	return NewRedisQueue(r.client, name, r.config), nil
}

// DeleteQueue deletes a queue
func (r *RedisBackend) DeleteQueue(ctx context.Context, name string) error {
	queue := NewRedisQueue(r.client, name, r.config)
	return queue.Clear(ctx)
}

// ListQueues lists all queues
func (r *RedisBackend) ListQueues(ctx context.Context) ([]string, error) {
	manager := NewRedisQueueManager(r.client, r.config)
	return manager.ListQueues(ctx)
}

// RabbitMQBackend implements QueueBackend using RabbitMQ
type RabbitMQBackend struct {
	conn   *amqp.Connection
	config *QueueConfig
}

// NewRabbitMQBackend creates a new RabbitMQ backend
func NewRabbitMQBackend(config *QueueConfig) (*RabbitMQBackend, error) {
	// Build connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &RabbitMQBackend{
		conn:   conn,
		config: config,
	}, nil
}

// Connect connects to RabbitMQ
func (r *RabbitMQBackend) Connect(ctx context.Context) error {
	// Connection is already established in NewRabbitMQBackend
	// Just verify it's still alive
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	return nil
}

// Disconnect disconnects from RabbitMQ
func (r *RabbitMQBackend) Disconnect() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// IsConnected returns true if connected
func (r *RabbitMQBackend) IsConnected() bool {
	ch, err := r.conn.Channel()
	if err != nil {
		return false
	}
	defer ch.Close()
	return true
}

// CreateQueue creates a queue
func (r *RabbitMQBackend) CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error) {
	return NewRabbitMQQueue(r.conn, name, r.config)
}

// DeleteQueue deletes a queue
func (r *RabbitMQBackend) DeleteQueue(ctx context.Context, name string) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDelete(name, false, false, false)
	return err
}

// ListQueues lists all queues
func (r *RabbitMQBackend) ListQueues(ctx context.Context) ([]string, error) {
	// RabbitMQ doesn't have a direct way to list queues via AMQP
	// This would typically require management API
	return []string{}, nil
}

// MemoryBackend implements QueueBackend using in-memory storage
type MemoryBackend struct {
	queues map[string]Queue
	config *QueueConfig
}

// NewMemoryBackend creates a new memory backend
func NewMemoryBackend(config *QueueConfig) (*MemoryBackend, error) {
	return &MemoryBackend{
		queues: make(map[string]Queue),
		config: config,
	}, nil
}

// Connect connects to memory backend (no-op)
func (m *MemoryBackend) Connect(ctx context.Context) error {
	return nil
}

// Disconnect disconnects from memory backend (no-op)
func (m *MemoryBackend) Disconnect() error {
	return nil
}

// IsConnected returns true if connected (always true for memory)
func (m *MemoryBackend) IsConnected() bool {
	return true
}

// CreateQueue creates a queue
func (m *MemoryBackend) CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error) {
	// Memory queue implementation would go here
	// For now, return an error as it's not implemented
	return nil, fmt.Errorf("memory backend not implemented")
}

// DeleteQueue deletes a queue
func (m *MemoryBackend) DeleteQueue(ctx context.Context, name string) error {
	delete(m.queues, name)
	return nil
}

// ListQueues lists all queues
func (m *MemoryBackend) ListQueues(ctx context.Context) ([]string, error) {
	queues := make([]string, 0, len(m.queues))
	for name := range m.queues {
		queues = append(queues, name)
	}
	return queues, nil
}
