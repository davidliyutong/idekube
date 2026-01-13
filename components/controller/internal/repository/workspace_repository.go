package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkspaceRepository handles workspace data access
type WorkspaceRepository struct {
	db *gorm.DB
}

// NewWorkspaceRepository creates a new workspace repository
func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *models.Workspace) error {
	return r.db.WithContext(ctx).Create(workspace).Error
}

// GetByID retrieves a workspace by ID
func (r *WorkspaceRepository) GetByID(ctx context.Context, id int64) (*models.Workspace, error) {
	var workspace models.Workspace
	err := r.db.WithContext(ctx).First(&workspace, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("workspace not found")
	}
	return &workspace, err
}

// GetByUUID retrieves a workspace by UUID
func (r *WorkspaceRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Workspace, error) {
	var workspace models.Workspace
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&workspace).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("workspace not found")
	}
	return &workspace, err
}

// Update updates a workspace
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *models.Workspace) error {
	return r.db.WithContext(ctx).Save(workspace).Error
}

// Delete deletes a workspace
func (r *WorkspaceRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Workspace{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("workspace not found")
	}
	return nil
}

// ListByOwner lists workspaces owned by a specific owner
func (r *WorkspaceRepository) ListByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) ([]*models.Workspace, error) {
	var workspaces []*models.Workspace
	err := r.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
		Order("created_at DESC").
		Find(&workspaces).Error
	return workspaces, err
}

// AttachVolume attaches a volume to a workspace
func (r *WorkspaceRepository) AttachVolume(ctx context.Context, wv *models.WorkspaceVolume) error {
	return r.db.WithContext(ctx).Create(wv).Error
}

// DetachVolume detaches a volume from a workspace
func (r *WorkspaceRepository) DetachVolume(ctx context.Context, workspaceID, volumeID int64) error {
	result := r.db.WithContext(ctx).Where("workspace_id = ? AND volume_id = ?", workspaceID, volumeID).
		Delete(&models.WorkspaceVolume{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("volume attachment not found")
	}
	return nil
}

// ListVolumes lists all volumes attached to a workspace
func (r *WorkspaceRepository) ListVolumes(ctx context.Context, workspaceID int64) ([]*models.WorkspaceVolume, error) {
	var volumes []*models.WorkspaceVolume
	err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&volumes).Error
	return volumes, err
}

// UpdateStatus updates workspace status
func (r *WorkspaceRepository) UpdateStatus(ctx context.Context, id int64, currentStatus, targetStatus models.WorkspaceStatus) error {
	return r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_status": currentStatus,
			"target_status":  targetStatus,
		}).Error
}
