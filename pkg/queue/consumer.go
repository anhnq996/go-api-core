package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConsumerImpl implements Consumer
type ConsumerImpl struct {
	queue   Queue
	handler MessageHandler
	options *ConsumerOptions
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// NewConsumer creates a new consumer
func NewConsumer(queue Queue, handler MessageHandler, options *ConsumerOptions) *ConsumerImpl {
	if options == nil {
		options = &ConsumerOptions{
			AutoAck:     false,
			Concurrency: 1,
			MaxRetries:  3,
			RetryDelay:  5 * time.Second,
		}
	}

	return &ConsumerImpl{
		queue:   queue,
		handler: handler,
		options: options,
	}
}

// Start starts consuming messages from the queue
func (c *ConsumerImpl) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("consumer is already running")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.running = true

	// Start worker goroutines
	for i := 0; i < c.options.Concurrency; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	return nil
}

// Stop stops consuming messages
func (c *ConsumerImpl) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return fmt.Errorf("consumer is not running")
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.wg.Wait()
	c.running = false

	return nil
}

// IsRunning returns true if the consumer is running
func (c *ConsumerImpl) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// GetQueue returns the queue being consumed
func (c *ConsumerImpl) GetQueue() Queue {
	return c.queue
}

// GetHandler returns the message handler
func (c *ConsumerImpl) GetHandler() MessageHandler {
	return c.handler
}

// worker processes messages from the queue
func (c *ConsumerImpl) worker(workerID int) {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Pop message from queue
			message, err := c.queue.PopWithTimeout(c.ctx, 1*time.Second)
			if err != nil {
				// Log error but continue
				continue
			}

			if message == nil {
				// No message, continue
				continue
			}

			// Process message
			c.processMessage(message)
		}
	}
}

// processMessage processes a single message
func (c *ConsumerImpl) processMessage(message *Message) {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	var err error
	retryCount := 0

	for retryCount <= c.options.MaxRetries {
		// Process message
		err = c.handler.Handle(ctx, message)
		if err == nil {
			// Message processed successfully
			return
		}

		// Handle error
		if c.handler.OnError != nil {
			if handleErr := c.handler.OnError(ctx, message, err); handleErr != nil {
				// If error handler fails, stop retrying
				return
			}
		}

		retryCount++
		message.RetryCount = retryCount

		// If we have retries left, wait before retrying
		if retryCount <= c.options.MaxRetries {
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(c.options.RetryDelay):
				continue
			}
		}
	}

	// Message failed after all retries
	// Could implement dead letter queue here
}

// ProducerImpl implements Producer
type ProducerImpl struct {
	queue Queue
}

// NewProducer creates a new producer
func NewProducer(queue Queue) *ProducerImpl {
	return &ProducerImpl{
		queue: queue,
	}
}

// Publish publishes a message to the queue
func (p *ProducerImpl) Publish(ctx context.Context, message *Message) error {
	return p.queue.Push(ctx, message)
}

// PublishBatch publishes multiple messages to the queue
func (p *ProducerImpl) PublishBatch(ctx context.Context, messages []*Message) error {
	for _, message := range messages {
		if err := p.queue.Push(ctx, message); err != nil {
			return fmt.Errorf("failed to publish message %s: %w", message.ID, err)
		}
	}
	return nil
}

// Close closes the producer connection
func (p *ProducerImpl) Close() error {
	return p.queue.Close()
}

// GetQueue returns the queue being published to
func (p *ProducerImpl) GetQueue() Queue {
	return p.queue
}

// DefaultMessageHandler provides a default message handler
type DefaultMessageHandler struct {
	ProcessFunc func(ctx context.Context, message *Message) error
	ErrorFunc   func(ctx context.Context, message *Message, err error) error
}

// Handle processes a message
func (h *DefaultMessageHandler) Handle(ctx context.Context, message *Message) error {
	if h.ProcessFunc != nil {
		return h.ProcessFunc(ctx, message)
	}
	return fmt.Errorf("no process function provided")
}

// OnError handles errors during message processing
func (h *DefaultMessageHandler) OnError(ctx context.Context, message *Message, err error) error {
	if h.ErrorFunc != nil {
		return h.ErrorFunc(ctx, message, err)
	}

	// Default error handling: log and continue
	fmt.Printf("Error processing message %s: %v\n", message.ID, err)
	return nil
}

// JobMessageHandler handles messages that implement the Job interface
type JobMessageHandler struct {
	JobFactory func(data []byte) (Job, error)
}

// Handle processes a message by creating and executing a job
func (h *JobMessageHandler) Handle(ctx context.Context, message *Message) error {
	job, err := h.JobFactory(message.Data)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return job.Process(ctx)
}

// OnError handles errors during job processing
func (h *JobMessageHandler) OnError(ctx context.Context, message *Message, err error) error {
	// Default error handling: log and continue
	fmt.Printf("Error processing job message %s: %v\n", message.ID, err)
	return nil
}
