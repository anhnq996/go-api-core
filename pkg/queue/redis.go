package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisQueue implements Queue using Redis
type RedisQueue struct {
	client *redis.Client
	name   string
	config *QueueConfig
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(client *redis.Client, name string, config *QueueConfig) *RedisQueue {
	return &RedisQueue{
		client: client,
		name:   name,
		config: config,
	}
}

// Push adds a message to the queue
func (r *RedisQueue) Push(ctx context.Context, message *Message) error {
	// Set message timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add to Redis list
	key := r.getQueueKey()
	if message.Delay > 0 {
		// Use delayed queue
		score := float64(time.Now().Add(message.Delay).Unix())
		return r.client.ZAdd(ctx, r.getDelayedQueueKey(), &redis.Z{
			Score:  score,
			Member: data,
		}).Err()
	}

	return r.client.LPush(ctx, key, data).Err()
}

// Pop retrieves and removes a message from the queue
func (r *RedisQueue) Pop(ctx context.Context) (*Message, error) {
	// First check delayed queue
	delayedKey := r.getDelayedQueueKey()
	now := float64(time.Now().Unix())

	// Get messages that are ready to be processed
	results, err := r.client.ZRangeByScoreWithScores(ctx, delayedKey, &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%f", now),
		Offset: 0,
		Count:  1,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to check delayed queue: %w", err)
	}

	if len(results) > 0 {
		// Remove from delayed queue
		r.client.ZRem(ctx, delayedKey, results[0].Member)

		// Parse message
		var message Message
		if err := json.Unmarshal([]byte(results[0].Member.(string)), &message); err != nil {
			return nil, fmt.Errorf("failed to unmarshal delayed message: %w", err)
		}

		return &message, nil
	}

	// Check regular queue
	key := r.getQueueKey()
	result := r.client.BRPop(ctx, 0, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, nil // No messages
		}
		return nil, fmt.Errorf("failed to pop message: %w", result.Err())
	}

	// Parse message
	var message Message
	if err := json.Unmarshal([]byte(result.Val()[1]), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// PopWithTimeout retrieves a message with timeout
func (r *RedisQueue) PopWithTimeout(ctx context.Context, timeout time.Duration) (*Message, error) {
	// First check delayed queue
	delayedKey := r.getDelayedQueueKey()
	now := float64(time.Now().Unix())

	// Get messages that are ready to be processed
	results, err := r.client.ZRangeByScoreWithScores(ctx, delayedKey, &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%f", now),
		Offset: 0,
		Count:  1,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to check delayed queue: %w", err)
	}

	if len(results) > 0 {
		// Remove from delayed queue
		r.client.ZRem(ctx, delayedKey, results[0].Member)

		// Parse message
		var message Message
		if err := json.Unmarshal([]byte(results[0].Member.(string)), &message); err != nil {
			return nil, fmt.Errorf("failed to unmarshal delayed message: %w", err)
		}

		return &message, nil
	}

	// Check regular queue with timeout
	key := r.getQueueKey()
	result := r.client.BRPop(ctx, timeout, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, nil // No messages
		}
		return nil, fmt.Errorf("failed to pop message: %w", result.Err())
	}

	// Parse message
	var message Message
	if err := json.Unmarshal([]byte(result.Val()[1]), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// Peek retrieves a message without removing it
func (r *RedisQueue) Peek(ctx context.Context) (*Message, error) {
	key := r.getQueueKey()
	result := r.client.LIndex(ctx, key, -1)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, nil // No messages
		}
		return nil, fmt.Errorf("failed to peek message: %w", result.Err())
	}

	// Parse message
	var message Message
	if err := json.Unmarshal([]byte(result.Val()), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// Size returns the number of messages in the queue
func (r *RedisQueue) Size(ctx context.Context) (int64, error) {
	key := r.getQueueKey()
	delayedKey := r.getDelayedQueueKey()

	// Get size of regular queue
	regularSize, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get regular queue size: %w", err)
	}

	// Get size of delayed queue
	delayedSize, err := r.client.ZCard(ctx, delayedKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get delayed queue size: %w", err)
	}

	return regularSize + delayedSize, nil
}

// Clear removes all messages from the queue
func (r *RedisQueue) Clear(ctx context.Context) error {
	key := r.getQueueKey()
	delayedKey := r.getDelayedQueueKey()

	// Clear regular queue
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear regular queue: %w", err)
	}

	// Clear delayed queue
	if err := r.client.Del(ctx, delayedKey).Err(); err != nil {
		return fmt.Errorf("failed to clear delayed queue: %w", err)
	}

	return nil
}

// Close closes the queue connection
func (r *RedisQueue) Close() error {
	// Redis client is managed externally, so we don't close it here
	return nil
}

// GetName returns the queue name
func (r *RedisQueue) GetName() string {
	return r.name
}

// getQueueKey returns the Redis key for the queue
func (r *RedisQueue) getQueueKey() string {
	return fmt.Sprintf("queue:%s", r.name)
}

// getDelayedQueueKey returns the Redis key for the delayed queue
func (r *RedisQueue) getDelayedQueueKey() string {
	return fmt.Sprintf("queue:%s:delayed", r.name)
}

// RedisQueueManager implements QueueManager using Redis
type RedisQueueManager struct {
	client *redis.Client
	config *QueueConfig
	queues map[string]Queue
}

// NewRedisQueueManager creates a new Redis queue manager
func NewRedisQueueManager(client *redis.Client, config *QueueConfig) *RedisQueueManager {
	return &RedisQueueManager{
		client: client,
		config: config,
		queues: make(map[string]Queue),
	}
}

// CreateQueue creates a new queue
func (r *RedisQueueManager) CreateQueue(ctx context.Context, name string, options *QueueOptions) (Queue, error) {
	if options == nil {
		options = r.config.DefaultQueueOptions
	}

	queue := NewRedisQueue(r.client, name, r.config)
	r.queues[name] = queue

	return queue, nil
}

// GetQueue returns an existing queue
func (r *RedisQueueManager) GetQueue(name string) (Queue, error) {
	if queue, exists := r.queues[name]; exists {
		return queue, nil
	}

	// Create queue if it doesn't exist
	return r.CreateQueue(context.Background(), name, nil)
}

// DeleteQueue deletes a queue
func (r *RedisQueueManager) DeleteQueue(ctx context.Context, name string) error {
	queue := NewRedisQueue(r.client, name, r.config)
	if err := queue.Clear(ctx); err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	delete(r.queues, name)
	return nil
}

// ListQueues returns a list of all queues
func (r *RedisQueueManager) ListQueues(ctx context.Context) ([]string, error) {
	pattern := "queue:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list queue keys: %w", err)
	}

	queues := make([]string, 0)
	seen := make(map[string]bool)

	for _, key := range keys {
		// Extract queue name from key
		if len(key) > 6 && key[:6] == "queue:" {
			queueName := key[6:]
			// Remove delayed queue suffix
			if len(queueName) > 9 && queueName[len(queueName)-9:] == ":delayed" {
				queueName = queueName[:len(queueName)-9]
			}

			if !seen[queueName] {
				queues = append(queues, queueName)
				seen[queueName] = true
			}
		}
	}

	return queues, nil
}

// Close closes all queue connections
func (r *RedisQueueManager) Close() error {
	// Redis client is managed externally, so we don't close it here
	return nil
}
