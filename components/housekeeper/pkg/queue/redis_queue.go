package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-housekeeper/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisQueueClient wraps Redis client for message queue operations
type RedisQueueClient struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisQueueClient creates a new Redis queue client
func NewRedisQueueClient(redisAddr, redisPassword string, redisDB int, log *logger.Logger) (*RedisQueueClient, error) {
	if log == nil {
		return nil, fmt.Errorf("logger is required")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueueClient{
		client: rdb,
		logger: log,
	}, nil
}

// Publish publishes a message to a Redis stream
func (c *RedisQueueClient) Publish(ctx context.Context, stream string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Use Redis Streams for message queue functionality
	_, err = c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"data":      string(body),
			"timestamp": time.Now().Unix(),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish to stream %s: %w", stream, err)
	}

	c.logger.Debugf("Published message to Redis stream: %s (size: %d)", stream, len(body))
	return nil
}

// Subscribe subscribes to a Redis stream and processes messages
func (c *RedisQueueClient) Subscribe(ctx context.Context, stream, consumerGroup, consumerName string, handler func([]byte) error) error {
	// Create consumer group if it doesn't exist
	// Ignore error if group already exists
	c.client.XGroupCreateMkStream(ctx, stream, consumerGroup, "0")

	c.logger.Infof("Starting Redis stream consumer - stream: %s, group: %s, consumer: %s", stream, consumerGroup, consumerName)

	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("Redis consumer stopped: %s", stream)
			return nil
		default:
			// Read messages from the stream
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    consumerGroup,
				Consumer: consumerName,
				Streams:  []string{stream, ">"},
				Count:    10,
				Block:    time.Second * 5,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// No new messages, continue
					continue
				}
				c.logger.Errorf("Failed to read from stream %s: %v", stream, err)
				time.Sleep(time.Second)
				continue
			}

			// Process messages
			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := c.processMessage(ctx, stream.Stream, consumerGroup, message, handler); err != nil {
						c.logger.Errorf("Failed to process message %s from stream %s: %v", message.ID, stream.Stream, err)
					}
				}
			}
		}
	}
}

// processMessage processes a single message
func (c *RedisQueueClient) processMessage(ctx context.Context, stream, consumerGroup string, message redis.XMessage, handler func([]byte) error) error {
	data, ok := message.Values["data"].(string)
	if !ok {
		c.logger.Errorf("Invalid message format: %s", message.ID)
		// Acknowledge malformed message to remove it from pending
		c.client.XAck(ctx, stream, consumerGroup, message.ID)
		return fmt.Errorf("invalid message format")
	}

	// Call handler
	if err := handler([]byte(data)); err != nil {
		c.logger.Errorf("Handler failed to process message %s: %v", message.ID, err)
		// Don't acknowledge - message will be retried
		return err
	}

	// Acknowledge message
	if err := c.client.XAck(ctx, stream, consumerGroup, message.ID).Err(); err != nil {
		c.logger.Errorf("Failed to acknowledge message %s: %v", message.ID, err)
		return err
	}

	c.logger.Debugf("Message processed successfully: %s", message.ID)
	return nil
}

// PublishStatusChange publishes a status change event to controller
func (c *RedisQueueClient) PublishStatusChange(ctx context.Context, event *StatusChangeEvent) error {
	return c.Publish(ctx, StreamControllerStatus, event)
}

// Close closes the Redis client
func (c *RedisQueueClient) Close() error {
	return c.client.Close()
}
