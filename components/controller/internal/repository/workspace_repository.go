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

// GetByID retrieves a workspace by ID (excludes soft deleted)
func (r *WorkspaceRepository) GetByID(ctx context.Context, id int64) (*models.Workspace, error) {
	var workspace models.Workspace
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&workspace, id).Error
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

// Delete soft deletes a workspace by setting deleted_at
func (r *WorkspaceRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&models.Workspace{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("workspace not found or already deleted")
	}
	return nil
}

// ListByUser lists workspaces owned by a specific user (excludes soft deleted)
func (r *WorkspaceRepository) ListByUser(ctx context.Context, userID int64) ([]*models.Workspace, error) {
	var workspaces []*models.Workspace
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).
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
func (r *WorkspaceRepository) UpdateStatus(ctx context.Context, id int64, currentStatus, targetStatus string) error {
	return r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_status": currentStatus,
			"target_status":  targetStatus,
		}).Error
}

// ListAll retrieves all workspaces (for admin) with pagination
func (r *WorkspaceRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Workspace{}).Where("deleted_at IS NULL")

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get workspaces with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&workspaces).Error

	if err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

// ListByOrganizationAll lists all workspaces in organizations where user is owner/admin
func (r *WorkspaceRepository) ListByOrganizationAll(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64

	// Get organizations where user is owner or admin
	query := r.db.WithContext(ctx).Model(&models.Workspace{}).
		Joins("INNER JOIN organizations ON workspaces.organization_id = organizations.id").
		Joins("INNER JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ? AND organization_members.role IN ? AND workspaces.deleted_at IS NULL", userID, []string{"owner", "admin"})

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("workspaces.identifier ILIKE ? OR workspaces.display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get workspaces with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("workspaces.%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&workspaces).Error

	if err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

// ListAccessibleByUser lists workspaces accessible by user (own + org member)
func (r *WorkspaceRepository) ListAccessibleByUser(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64

	// Workspaces owned by user OR in organizations where user is a member
	query := r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("(user_id = ? AND deleted_at IS NULL)", userID).
		Or("organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = ?) AND deleted_at IS NULL", userID)

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get workspaces with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&workspaces).Error

	if err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

// ListByLabel lists workspaces by label filters
func (r *WorkspaceRepository) ListByLabel(ctx context.Context, labels map[string]string, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Workspace{}).Where("deleted_at IS NULL")

	// Apply label filters
	for key, value := range labels {
		query = query.Where("labels->>? = ?", key, value)
	}

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get workspaces with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&workspaces).Error

	if err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

// UpdateLabels updates workspace labels
func (r *WorkspaceRepository) UpdateLabels(ctx context.Context, id int64, labels models.ResourceLabels) error {
	return r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("id = ?", id).
		Update("labels", labels).Error
}

// ListPublicInOrganization lists public workspaces in a specific organization
func (r *WorkspaceRepository) ListPublicInOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("organization_id = ? AND is_public = ? AND deleted_at IS NULL", orgID, true)

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get workspaces with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&workspaces).Error

	if err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}

// UpdateIsPublic updates the is_public flag of a workspace
func (r *WorkspaceRepository) UpdateIsPublic(ctx context.Context, id int64, isPublic bool) error {
	return r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("id = ?", id).
		Update("is_public", isPublic).Error
}

// ListByOrganization lists workspaces in a specific organization
func (r *WorkspaceRepository) ListByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Workspace{}).
		Where("organization_id = ? AND deleted_at IS NULL", orgID)

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get workspaces with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&workspaces).Error

	if err != nil {
		return nil, 0, err
	}

	return workspaces, total, nil
}
