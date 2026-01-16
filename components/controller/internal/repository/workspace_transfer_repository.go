package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// WorkspaceTransferRepository handles workspace transfer database operations
type WorkspaceTransferRepository struct {
	db *gorm.DB
}

// NewWorkspaceTransferRepository creates a new workspace transfer repository
func NewWorkspaceTransferRepository(db *gorm.DB) *WorkspaceTransferRepository {
	return &WorkspaceTransferRepository{db: db}
}

// Create creates a new workspace transfer request
func (r *WorkspaceTransferRepository) Create(ctx context.Context, transfer *models.WorkspaceTransfer) error {
	result := r.db.WithContext(ctx).Create(transfer)
	return result.Error
}

// GetByID retrieves a workspace transfer by ID
func (r *WorkspaceTransferRepository) GetByID(ctx context.Context, id int64) (*models.WorkspaceTransfer, error) {
	var transfer models.WorkspaceTransfer
	result := r.db.WithContext(ctx).First(&transfer, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workspace transfer not found")
		}
		return nil, result.Error
	}
	return &transfer, nil
}

// ListPendingForUser retrieves pending transfer requests for a user (as recipient)
func (r *WorkspaceTransferRepository) ListPendingForUser(ctx context.Context, userID int64) ([]*models.WorkspaceTransfer, error) {
	var transfers []*models.WorkspaceTransfer
	result := r.db.WithContext(ctx).
		Where("to_user_id = ? AND status = ?", userID, "pending").
		Order("created_at DESC").
		Find(&transfers)
	if result.Error != nil {
		return nil, result.Error
	}
	return transfers, nil
}

// ListByWorkspace retrieves all transfers for a workspace
func (r *WorkspaceTransferRepository) ListByWorkspace(ctx context.Context, workspaceID int64) ([]*models.WorkspaceTransfer, error) {
	var transfers []*models.WorkspaceTransfer
	result := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&transfers)
	if result.Error != nil {
		return nil, result.Error
	}
	return transfers, nil
}

// Update updates a workspace transfer
func (r *WorkspaceTransferRepository) Update(ctx context.Context, transfer *models.WorkspaceTransfer) error {
	result := r.db.WithContext(ctx).Save(transfer)
	return result.Error
}

// HasPendingTransfer checks if a workspace has any pending transfer
func (r *WorkspaceTransferRepository) HasPendingTransfer(ctx context.Context, workspaceID int64) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.WorkspaceTransfer{}).
		Where("workspace_id = ? AND status = ?", workspaceID, "pending").
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// CancelPendingTransfers cancels all pending transfers for a workspace
func (r *WorkspaceTransferRepository) CancelPendingTransfers(ctx context.Context, workspaceID int64) error {
	result := r.db.WithContext(ctx).
		Model(&models.WorkspaceTransfer{}).
		Where("workspace_id = ? AND status = ?", workspaceID, "pending").
		Update("status", "cancelled")
	return result.Error
}
