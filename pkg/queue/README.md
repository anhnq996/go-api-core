# Queue Package

A flexible message queue system with support for Redis and RabbitMQ backends.

## Features

- **Multiple Backends**: Redis and RabbitMQ support
- **Easy Switching**: Change backends without code changes
- **Message Persistence**: Durable message storage
- **Retry Mechanism**: Configurable retry logic
- **Dead Letter Queue**: Handle failed messages
- **Priority Queues**: Message prioritization
- **Delayed Messages**: Schedule messages for future processing
- **Batch Operations**: Efficient batch publishing
- **Consumer Management**: Multiple concurrent workers
- **Connection Pooling**: Optimized connection management

## Installation

```bash
go get github.com/go-redis/redis/v8
go get github.com/streadway/amqp
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "your-project/pkg/queue"
    "github.com/go-redis/redis/v8"
)

func main() {
    // Create queue manager
    config := &queue.QueueConfig{
        Type: queue.QueueTypeRedis,
        Host: "localhost",
        Port: 6379,
    }

    manager, err := queue.NewQueueManager(config)
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()

    // Create queue
    queue, err := manager.CreateQueue(context.Background(), "my-queue", nil)
    if err != nil {
        log.Fatal(err)
    }

    // Publish message
    message := &queue.Message{
        ID:   "msg-1",
        Data: []byte("Hello, World!"),
        Headers: map[string]string{
            "type": "greeting",
        },
    }

    err = queue.Push(context.Background(), message)
    if err != nil {
        log.Fatal(err)
    }

    // Consume message
    consumer := queue.NewConsumer(queue, &MyHandler{}, nil)
    consumer.Start(context.Background())

    // Keep running
    select {}
}

type MyHandler struct{}

func (h *MyHandler) Handle(ctx context.Context, message *queue.Message) error {
    fmt.Printf("Received: %s\n", string(message.Data))
    return nil
}

func (h *MyHandler) OnError(ctx context.Context, message *queue.Message, err error) error {
    fmt.Printf("Error processing message: %v\n", err)
    return nil
}
```

### Switching Between Backends

```go
// Redis Configuration
redisConfig := &queue.QueueConfig{
    Type: queue.QueueTypeRedis,
    Host: "localhost",
    Port: 6379,
    Database: 0,
}

// RabbitMQ Configuration
rabbitConfig := &queue.QueueConfig{
    Type: queue.QueueTypeRabbitMQ,
    Host: "localhost",
    Port: 5672,
    Username: "guest",
    Password: "guest",
    VHost: "/",
}

// Create manager with desired backend
manager, err := queue.NewQueueManager(redisConfig) // or rabbitConfig
```

## Configuration

### Queue Config

```go
type QueueConfig struct {
    Type     QueueType // redis, rabbitmq, memory
    Host     string
    Port     int
    Username string
    Password string
    Database int       // Redis only
    VHost    string    // RabbitMQ only

    // Connection options
    MaxRetries      int
    RetryDelay      time.Duration
    ConnectTimeout  time.Duration
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    IdleTimeout     time.Duration
    MaxIdleConns    int
    MaxActiveConns  int
}
```

### Queue Options

```go
type QueueOptions struct {
    Durable              bool                   // Survive broker restarts
    AutoDelete           bool                   // Delete when unused
    Exclusive            bool                   // Exclusive to connection
    NoWait               bool                   // Don't wait for confirmation
    Arguments            map[string]interface{} // Additional arguments
    MaxLength            int64                  // Max messages in queue
    MessageTTL           time.Duration          // Message TTL
    DeadLetterExchange   string                 // Dead letter exchange
    DeadLetterRoutingKey string                 // Dead letter routing key
}
```

### Consumer Options

```go
type ConsumerOptions struct {
    AutoAck     bool          // Auto acknowledge messages
    Exclusive   bool          // Exclusive consumer
    NoLocal     bool          // Don't receive own messages
    NoWait      bool          // Don't wait for confirmation
    PrefetchCount int         // Number of messages to prefetch
    PrefetchSize  int         // Prefetch size in bytes
    Global      bool          // Global prefetch
    Concurrency int           // Number of concurrent workers
    RetryDelay  time.Duration // Delay between retries
    MaxRetries  int           // Maximum retries
}
```

## Message Structure

```go
type Message struct {
    ID         string            // Unique message ID
    Data       []byte            // Message payload
    Headers    map[string]string // Message headers
    Timestamp  time.Time         // Message timestamp
    RetryCount int               // Current retry count
    MaxRetries int               // Maximum retries
    Delay      time.Duration     // Delay before processing
    Priority   int               // Message priority
}
```

## Advanced Usage

### Producer with Batch Publishing

```go
func publishBatch(producer queue.Producer, messages []string) error {
    queueMessages := make([]*queue.Message, len(messages))
    for i, msg := range messages {
        queueMessages[i] = &queue.Message{
            ID:   fmt.Sprintf("msg-%d", i),
            Data: []byte(msg),
        }
    }

    return producer.PublishBatch(context.Background(), queueMessages)
}
```

### Consumer with Custom Handler

```go
type EmailHandler struct {
    emailService EmailService
}

func (h *EmailHandler) Handle(ctx context.Context, message *queue.Message) error {
    var email Email
    if err := json.Unmarshal(message.Data, &email); err != nil {
        return fmt.Errorf("failed to unmarshal email: %w", err)
    }

    return h.emailService.Send(ctx, email)
}

func (h *EmailHandler) OnError(ctx context.Context, message *queue.Message, err error) error {
    // Log error and potentially send to dead letter queue
    log.Printf("Failed to send email: %v", err)

    // Could implement dead letter queue logic here
    return nil
}
```

### Priority Queue

```go
// High priority message
highPriorityMsg := &queue.Message{
    ID:       "urgent-1",
    Data:     []byte("urgent data"),
    Priority: 10, // Higher number = higher priority
}

// Low priority message
lowPriorityMsg := &queue.Message{
    ID:       "normal-1",
    Data:     []byte("normal data"),
    Priority: 1,
}
```

### Delayed Messages

```go
// Message to be processed after 1 hour
delayedMsg := &queue.Message{
    ID:    "delayed-1",
    Data:  []byte("delayed data"),
    Delay: 1 * time.Hour,
}
```

### Dead Letter Queue Setup

```go
// Create main queue with dead letter configuration
options := &queue.QueueOptions{
    Durable: true,
    Arguments: map[string]interface{}{
        "x-dead-letter-exchange": "dlx",
        "x-dead-letter-routing-key": "failed",
    },
}

mainQueue, err := manager.CreateQueue(ctx, "main-queue", options)

// Create dead letter queue
dlqOptions := &queue.QueueOptions{
    Durable: true,
}
dlq, err := manager.CreateQueue(ctx, "dead-letter-queue", dlqOptions)
```

## Backend-Specific Features

### Redis Features

- **Delayed Messages**: Using sorted sets
- **Priority Queues**: Using sorted sets
- **Persistence**: RDB and AOF support
- **Clustering**: Redis Cluster support
- **Pub/Sub**: Real-time messaging

### RabbitMQ Features

- **Exchanges**: Direct, topic, fanout, headers
- **Routing**: Flexible message routing
- **Dead Letter Exchanges**: Built-in DLX support
- **Message TTL**: Per-message and per-queue TTL
- **Clustering**: High availability clustering

## Error Handling

```go
type RetryHandler struct {
    maxRetries int
    retryDelay time.Duration
}

func (h *RetryHandler) Handle(ctx context.Context, message *queue.Message) error {
    // Your processing logic
    return processMessage(message)
}

func (h *RetryHandler) OnError(ctx context.Context, message *queue.Message, err error) error {
    if message.RetryCount < h.maxRetries {
        // Retry with exponential backoff
        delay := h.retryDelay * time.Duration(1<<message.RetryCount)
        time.Sleep(delay)
        return err // Return error to retry
    }

    // Max retries exceeded, send to dead letter queue
    return sendToDeadLetterQueue(ctx, message, err)
}
```

## Monitoring

### Queue Statistics

```go
// Get queue size
size, err := queue.Size(ctx)

// List all queues
queues, err := manager.ListQueues(ctx)

// Check consumer status
isRunning := consumer.IsRunning()
```

### Health Checks

```go
func healthCheck(manager queue.QueueManager) error {
    // Check if manager is connected
    if !manager.IsConnected() {
        return fmt.Errorf("queue manager not connected")
    }

    // Test queue operations
    testQueue, err := manager.GetQueue("health-check")
    if err != nil {
        return fmt.Errorf("failed to get test queue: %w", err)
    }

    // Test publish/consume
    testMsg := &queue.Message{
        ID:   "health-check",
        Data: []byte("test"),
    }

    if err := testQueue.Push(ctx, testMsg); err != nil {
        return fmt.Errorf("failed to publish test message: %w", err)
    }

    if _, err := testQueue.Pop(ctx); err != nil {
        return fmt.Errorf("failed to consume test message: %w", err)
    }

    return nil
}
```

## Best Practices

1. **Use appropriate backends** for your use case
2. **Set reasonable timeouts** for operations
3. **Handle errors gracefully** with retry logic
4. **Use dead letter queues** for failed messages
5. **Monitor queue sizes** to prevent memory issues
6. **Use batch operations** for better performance
7. **Implement proper logging** for debugging
8. **Use connection pooling** for high throughput
9. **Set up monitoring** and alerting
10. **Test with realistic load** before production

## Examples

See the `examples/` directory for complete examples:

- Basic producer/consumer
- Batch operations
- Error handling
- Dead letter queues
- Priority queues
- Delayed messages
- Multi-backend switching
