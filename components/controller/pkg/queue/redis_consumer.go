package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/davidliyutong/idekube-controller/internal/repository"
)

// StatusConsumer consumes status change events from HouseKeeper
type StatusConsumer struct {
	queue         *RedisQueueClient
	workspaceRepo *repository.WorkspaceRepository
	logger        *zap.Logger
}

// NewStatusConsumer creates a new status consumer
func NewStatusConsumer(
	queue *RedisQueueClient,
	workspaceRepo *repository.WorkspaceRepository,
	logger *zap.Logger,
) (*StatusConsumer, error) {
	if queue == nil {
		return nil, fmt.Errorf("redis queue client is required")
	}
	if workspaceRepo == nil {
		return nil, fmt.Errorf("workspace repository is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &StatusConsumer{
		queue:         queue,
		workspaceRepo: workspaceRepo,
		logger:        logger,
	}, nil
}

// Start starts consuming status change events
func (c *StatusConsumer) Start(ctx context.Context) error {
	c.logger.Info("Status consumer started", zap.String("stream", StreamControllerStatus))

	// Subscribe to the status stream
	return c.queue.Subscribe(
		ctx,
		StreamControllerStatus,
		ConsumerGroupControllerStatus,
		"controller-status-consumer",
		c.handleMessage,
	)
}

// handleMessage handles a single message
func (c *StatusConsumer) handleMessage(data []byte) error {
	// Parse event
	var event StatusChangeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		c.logger.Error("Failed to unmarshal status change event",
			zap.Error(err),
			zap.String("body", string(data)))
		return err
	}

	c.logger.Info("Received status change event",
		zap.Int64("workspace_id", event.WorkspaceID),
		zap.String("old_status", string(event.OldStatus)),
		zap.String("new_status", string(event.NewStatus)),
		zap.String("reason", event.Reason))

	// Handle event
	ctx := context.Background()
	if err := c.handleStatusChange(ctx, &event); err != nil {
		c.logger.Error("Failed to handle status change",
			zap.Int64("workspace_id", event.WorkspaceID),
			zap.Error(err))
		return err
	}

	return nil
}

// handleStatusChange processes a status change event
func (c *StatusConsumer) handleStatusChange(ctx context.Context, event *StatusChangeEvent) error {
	// Get workspace
	workspace, err := c.workspaceRepo.GetByID(ctx, event.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Update status
	if err := c.workspaceRepo.UpdateStatus(ctx, event.WorkspaceID, event.NewStatus, workspace.TargetStatus); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	// Update K8S resource information if provided
	if event.K8SResourceStatus != nil {
		// Could store this in a separate table or in workspace metadata
		c.logger.Info("K8S resource status updated",
			zap.Int64("workspace_id", event.WorkspaceID),
			zap.Bool("deployment_ready", event.K8SResourceStatus.DeploymentReady),
			zap.Int32("pods_ready", event.K8SResourceStatus.PodsReady),
			zap.Int32("pods_total", event.K8SResourceStatus.PodsTotal))
	}

	// Handle specific event types
	switch event.Type {
	case EventTypeTimeout:
		c.logger.Warn("Workspace timed out",
			zap.Int64("workspace_id", event.WorkspaceID),
			zap.String("reason", event.Reason))
		// Could send notification to user
	case EventTypeK8SError:
		c.logger.Error("K8S error for workspace",
			zap.Int64("workspace_id", event.WorkspaceID),
			zap.String("error", event.ErrorMessage))
		// Could send alert
	}

	return nil
}
