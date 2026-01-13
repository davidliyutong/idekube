package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/davidliyutong/idekube-controller/internal/models"
)

// EventPublisher publishes events to message queue
type EventPublisher struct {
	mq     *RabbitMQClient
	logger *zap.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(mq *RabbitMQClient, logger *zap.Logger) (*EventPublisher, error) {
	if mq == nil {
		return nil, fmt.Errorf("rabbitmq client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	publisher := &EventPublisher{
		mq:     mq,
		logger: logger,
	}

	// Initialize exchange and queues
	if err := publisher.setupTopology(); err != nil {
		return nil, fmt.Errorf("failed to setup topology: %w", err)
	}

	return publisher, nil
}

// setupTopology creates exchange and queues
func (p *EventPublisher) setupTopology() error {
	channel := p.mq.Channel()

	// Declare exchange
	if err := channel.ExchangeDeclare(
		ExchangeIDEKube, // name
		ExchangeType,    // type: topic
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queues
	queues := []struct {
		name       string
		routingKey string
	}{
		{QueueHousekeeperWorkspace, RoutingKeyWorkspace},
		{QueueHousekeeperVolume, RoutingKeyVolume},
		{QueueControllerStatus, RoutingKeyStatus},
	}

	for _, q := range queues {
		// Declare queue
		_, err := channel.QueueDeclare(
			q.name, // name
			true,   // durable
			false,  // delete when unused
			false,  // exclusive
			false,  // no-wait
			nil,    // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		// Bind queue to exchange
		if err := channel.QueueBind(
			q.name,          // queue name
			q.routingKey,    // routing key
			ExchangeIDEKube, // exchange
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q.name, err)
		}
	}

	p.logger.Info("Message queue topology setup completed",
		zap.String("exchange", ExchangeIDEKube),
		zap.Int("queues", len(queues)))

	return nil
}

// publish publishes an event to the exchange
func (p *EventPublisher) publish(ctx context.Context, routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := p.mq.Channel()

	// Create context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		publishCtx,
		ExchangeIDEKube, // exchange
		routingKey,      // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persistent messages
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// --- Workspace Events ---

// PublishWorkspaceCreate publishes a workspace creation event
func (p *EventPublisher) PublishWorkspaceCreate(ctx context.Context, workspace *models.Workspace, template *models.Template, volumes []*models.Volume) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceCreate, workspace, template, volumes)
	event.UserID = workspace.OwnerID

	if err := p.publish(ctx, EventTypeWorkspaceCreate, event); err != nil {
		p.logger.Error("Failed to publish workspace create event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace create event",
		zap.Int64("workspace_id", workspace.ID),
		zap.String("name", workspace.Name))

	return nil
}

// PublishWorkspaceUpdate publishes a workspace update event
func (p *EventPublisher) PublishWorkspaceUpdate(ctx context.Context, workspace *models.Workspace, template *models.Template, volumes []*models.Volume, reason string) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceUpdate, workspace, template, volumes)
	event.UserID = workspace.OwnerID
	event.Reason = reason

	if err := p.publish(ctx, EventTypeWorkspaceUpdate, event); err != nil {
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

	if err := p.publish(ctx, EventTypeWorkspaceDelete, event); err != nil {
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
func (p *EventPublisher) PublishWorkspaceStart(ctx context.Context, workspace *models.Workspace) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceStart, workspace, nil, nil)
	event.UserID = workspace.OwnerID

	if err := p.publish(ctx, EventTypeWorkspaceStart, event); err != nil {
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
func (p *EventPublisher) PublishWorkspaceStop(ctx context.Context, workspace *models.Workspace, reason string) error {
	event := NewWorkspaceEvent(EventTypeWorkspaceStop, workspace, nil, nil)
	event.UserID = workspace.OwnerID
	event.Reason = reason

	if err := p.publish(ctx, EventTypeWorkspaceStop, event); err != nil {
		p.logger.Error("Failed to publish workspace stop event",
			zap.Int64("workspace_id", workspace.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published workspace stop event",
		zap.Int64("workspace_id", workspace.ID),
		zap.String("reason", reason))

	return nil
}

// --- Volume Events ---

// PublishVolumeCreate publishes a volume creation event
func (p *EventPublisher) PublishVolumeCreate(ctx context.Context, volume *models.Volume) error {
	event := NewVolumeEvent(EventTypeVolumeCreate, volume)
	event.UserID = volume.OwnerID

	if err := p.publish(ctx, EventTypeVolumeCreate, event); err != nil {
		p.logger.Error("Failed to publish volume create event",
			zap.Int64("volume_id", volume.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published volume create event",
		zap.Int64("volume_id", volume.ID),
		zap.String("name", volume.Name),
		zap.Int("size_mb", volume.SizeMB))

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

	if err := p.publish(ctx, EventTypeVolumeDelete, event); err != nil {
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
func (p *EventPublisher) PublishVolumeResize(ctx context.Context, volume *models.Volume, oldSize string) error {
	event := NewVolumeEvent(EventTypeVolumeResize, volume)
	event.UserID = volume.OwnerID
	event.Reason = fmt.Sprintf("Resize from %s to %d MB", oldSize, volume.SizeMB)

	if err := p.publish(ctx, EventTypeVolumeResize, event); err != nil {
		p.logger.Error("Failed to publish volume resize event",
			zap.Int64("volume_id", volume.ID),
			zap.Error(err))
		return err
	}

	p.logger.Info("Published volume resize event",
		zap.Int64("volume_id", volume.ID),
		zap.String("old_size", oldSize),
		zap.Int("new_size_mb", volume.SizeMB))

	return nil
}

// Close closes the publisher (currently no-op, managed by RabbitMQClient)
func (p *EventPublisher) Close() error {
	// RabbitMQClient handles the actual connection closing
	return nil
}
