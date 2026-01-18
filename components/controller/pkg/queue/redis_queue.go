package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/davidliyutong/idekube-controller/internal/models"
)

// RedisQueueClient wraps Redis client for message queue operations
type RedisQueueClient struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisQueueClient creates a new Redis queue client
func NewRedisQueueClient(client *redis.Client, logger *zap.Logger) (*RedisQueueClient, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &RedisQueueClient{
		client: client,
		logger: logger,
	}, nil
}

// Publish publishes a message to a Redis stream
func (c *RedisQueueClient) Publish(ctx context.Context, stream string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Use Redis Streams for message queue functionality
	// XAdd adds a new entry to the stream
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

	c.logger.Debug("Published message to Redis stream",
		zap.String("stream", stream),
		zap.Int("size", len(body)))

	return nil
}

// Subscribe subscribes to a Redis stream and processes messages
func (c *RedisQueueClient) Subscribe(ctx context.Context, stream, consumerGroup, consumerName string, handler func([]byte) error) error {
	// Create consumer group if it doesn't exist
	// Ignore error if group already exists
	c.client.XGroupCreateMkStream(ctx, stream, consumerGroup, "0")

	c.logger.Info("Starting Redis stream consumer",
		zap.String("stream", stream),
		zap.String("consumer_group", consumerGroup),
		zap.String("consumer_name", consumerName))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Redis consumer stopped", zap.String("stream", stream))
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
				c.logger.Error("Failed to read from stream",
					zap.String("stream", stream),
					zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			// Process messages
			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := c.processMessage(ctx, stream.Stream, consumerGroup, message, handler); err != nil {
						c.logger.Error("Failed to process message",
							zap.String("stream", stream.Stream),
							zap.String("message_id", message.ID),
							zap.Error(err))
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
		c.logger.Error("Invalid message format", zap.String("message_id", message.ID))
		// Acknowledge malformed message to remove it from pending
		c.client.XAck(ctx, stream, consumerGroup, message.ID)
		return fmt.Errorf("invalid message format")
	}

	// Call handler
	if err := handler([]byte(data)); err != nil {
		c.logger.Error("Handler failed to process message",
			zap.String("message_id", message.ID),
			zap.Error(err))
		// Don't acknowledge - message will be retried
		return err
	}

	// Acknowledge message
	if err := c.client.XAck(ctx, stream, consumerGroup, message.ID).Err(); err != nil {
		c.logger.Error("Failed to acknowledge message",
			zap.String("message_id", message.ID),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Message processed successfully", zap.String("message_id", message.ID))
	return nil
}

// Close closes the Redis client
func (c *RedisQueueClient) Close() error {
	// Redis client is managed externally, no need to close here
	return nil
}

// EventPublisher publishes events to Redis streams
type EventPublisher struct {
	queue  *RedisQueueClient
	logger *zap.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(queue *RedisQueueClient, logger *zap.Logger) (*EventPublisher, error) {
	if queue == nil {
		return nil, fmt.Errorf("redis queue client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &EventPublisher{
		queue:  queue,
		logger: logger,
	}, nil
}

// publish publishes an event to the stream
func (p *EventPublisher) publish(ctx context.Context, stream string, event interface{}) error {
	if err := p.queue.Publish(ctx, stream, event); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

// --- Workspace Events ---

// PublishWorkspaceCreate publishes a workspace creation event
func (p *EventPublisher) PublishWorkspaceCreate(ctx context.Context, workspace *models.Workspace, template *models.Template, volumes []*models.Volume) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceCreate, workspace, template, volumes)
	event.UserID = workspace.CreatedBy

	if err := p.publish(ctx, StreamHousekeeperWorkspace, event); err != nil {
		p.logger.Error("Failed to publish workspace create event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace create event",
		zap.Int64("workspace_id", workspace.ID),
		zap.String("name", workspace.Identifier))

	return nil
}

// PublishWorkspaceUpdate publishes a workspace update event
func (p *EventPublisher) PublishWorkspaceUpdate(ctx context.Context, workspace *models.Workspace, template *models.Template, volumes []*models.Volume, reason string) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceUpdate, workspace, template, volumes)
	event.UserID = workspace.CreatedBy
	event.Reason = reason

	if err := p.publish(ctx, StreamHousekeeperWorkspace, event); err != nil {
		p.logger.Error("Failed to publish workspace update event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace update event",
		zap.Int64("workspace_id", workspace.ID),
		zap.String("reason", reason))

	return nil
}

// PublishWorkspaceDelete publishes a workspace deletion event
func (p *EventPublisher) PublishWorkspaceDelete(ctx context.Context, workspaceID int64, workspace *models.Workspace) error {
	event := &WorkspaceEvent{
		Type:        EventTypeWorkspaceDelete,
		WorkspaceID: workspaceID,
		Workspace:   workspace,
		Timestamp:   time.Now(),
	}

	if err := p.publish(ctx, StreamHousekeeperWorkspace, event); err != nil {
		p.logger.Error("Failed to publish workspace delete event",
			zap.Int64("workspace_id", workspaceID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace delete event",
		zap.Int64("workspace_id", workspaceID))

	return nil
}

// PublishWorkspaceStart publishes a workspace start event
func (p *EventPublisher) PublishWorkspaceStart(ctx context.Context, workspace *models.Workspace, template *models.Template, volumes []*models.Volume) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceStart, workspace, template, volumes)
	event.UserID = workspace.CreatedBy

	if err := p.publish(ctx, StreamHousekeeperWorkspace, event); err != nil {
		p.logger.Error("Failed to publish workspace start event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace start event",
		zap.Int64("workspace_id", workspace.ID))

	return nil
}

// PublishWorkspaceStop publishes a workspace stop event
func (p *EventPublisher) PublishWorkspaceStop(ctx context.Context, workspace *models.Workspace) error {
	event := &WorkspaceEvent{
		Type:        EventTypeWorkspaceStop,
		WorkspaceID: workspace.ID,
		Workspace:   workspace,
		Timestamp:   time.Now(),
		UserID:      workspace.CreatedBy,
	}

	if err := p.publish(ctx, StreamHousekeeperWorkspace, event); err != nil {
		p.logger.Error("Failed to publish workspace stop event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace stop event",
		zap.Int64("workspace_id", workspace.ID))

	return nil
}

// --- Volume Events ---

// PublishVolumeCreate publishes a volume creation event
func (p *EventPublisher) PublishVolumeCreate(ctx context.Context, volume *models.Volume) error {
	event := NewVolumeEvent(EventTypeVolumeCreate, volume)
	// Note: OwnerID is now Organization ID, not User ID
	// Ensure UserID is set so the housekeeper can track who initiated the creation
	event.UserID = volume.OwnerID

	if err := p.publish(ctx, StreamHousekeeperVolume, event); err != nil {
		p.logger.Error("Failed to publish volume create event",
			zap.Int64("volume_id", volume.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published volume create event",
		zap.Int64("volume_id", volume.ID),
		zap.String("name", volume.Identifier))

	return nil
}

// PublishVolumeDelete publishes a volume deletion event
func (p *EventPublisher) PublishVolumeDelete(ctx context.Context, volumeID int64, volume *models.Volume) error {
	event := &VolumeEvent{
		Type:      EventTypeVolumeDelete,
		VolumeID:  volumeID,
		Volume:    volume,
		Timestamp: time.Now(),
	}

	if err := p.publish(ctx, StreamHousekeeperVolume, event); err != nil {
		p.logger.Error("Failed to publish volume delete event",
			zap.Int64("volume_id", volumeID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published volume delete event",
		zap.Int64("volume_id", volumeID))

	return nil
}

// PublishVolumeResize publishes a volume resize event
func (p *EventPublisher) PublishVolumeResize(ctx context.Context, volume *models.Volume, oldSize, newSize int64) error {
	event := NewVolumeEvent(EventTypeVolumeResize, volume)
	// Note: OwnerID is now Organization ID, not User ID
	// event.UserID should be set by the caller if needed

	if err := p.publish(ctx, StreamHousekeeperVolume, event); err != nil {
		p.logger.Error("Failed to publish volume resize event",
			zap.Int64("volume_id", volume.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published volume resize event",
		zap.Int64("volume_id", volume.ID),
		zap.Int64("old_size", oldSize),
		zap.Int64("new_size", newSize))

	return nil
}

// --- User Events ---

// PublishUserDelete publishes a user deletion event
func (p *EventPublisher) PublishUserDelete(ctx context.Context, userID int64, username string) error {
	event := &UserDeleteEvent{
		Type:      EventTypeUserDelete,
		UserID:    userID,
		Username:  username,
		Timestamp: time.Now(),
	}

	if err := p.publish(ctx, StreamHousekeeperCleanup, event); err != nil {
		p.logger.Error("Failed to publish user delete event",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published user delete event",
		zap.Int64("user_id", userID),
		zap.String("username", username))

	return nil
}

// --- Organization Events ---

// PublishOrganizationDelete publishes an organization deletion event
func (p *EventPublisher) PublishOrganizationDelete(ctx context.Context, orgID int64, name string) error {
	event := &OrganizationDeleteEvent{
		Type:           EventTypeOrganizationDelete,
		OrganizationID: orgID,
		Name:           name,
		Timestamp:      time.Now(),
	}

	if err := p.publish(ctx, StreamHousekeeperCleanup, event); err != nil {
		p.logger.Error("Failed to publish organization delete event",
			zap.Int64("organization_id", orgID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published organization delete event",
		zap.Int64("organization_id", orgID),
		zap.String("name", name))

	return nil
}
