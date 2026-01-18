package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VolumeRepository handles volume data access
type VolumeRepository struct {
	db *gorm.DB
}

// NewVolumeRepository creates a new volume repository
func NewVolumeRepository(db *gorm.DB) *VolumeRepository {
	return &VolumeRepository{db: db}
}

// Create creates a new volume
func (r *VolumeRepository) Create(ctx context.Context, volume *models.Volume) error {
	return r.db.WithContext(ctx).Create(volume).Error
}

// GetByID retrieves a volume by ID (excludes soft deleted)
func (r *VolumeRepository) GetByID(ctx context.Context, id int64) (*models.Volume, error) {
	var volume models.Volume
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&volume, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("volume not found")
	}
	return &volume, err
}

// GetByUUID retrieves a volume by UUID
func (r *VolumeRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Volume, error) {
	var volume models.Volume
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&volume).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("volume not found")
	}
	return &volume, err
}

// Update updates a volume
func (r *VolumeRepository) Update(ctx context.Context, volume *models.Volume) error {
	return r.db.WithContext(ctx).Save(volume).Error
}

// Delete soft deletes a volume by setting deleted_at
func (r *VolumeRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&models.Volume{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("volume not found or already deleted")
	}
	return nil
}

// ListByUser lists volumes owned by a specific user (excludes soft deleted)
func (r *VolumeRepository) ListByUser(ctx context.Context, userID int64) ([]*models.Volume, error) {
	var volumes []*models.Volume
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&volumes).Error
	return volumes, err
}

// UpdateStatus updates volume status
func (r *VolumeRepository) UpdateStatus(ctx context.Context, id int64, status string, pvcName *string) error {
	return r.db.WithContext(ctx).Model(&models.Volume{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"k8s_pvc_name": pvcName,
		}).Error
}

// ListAll retrieves all volumes (for admin) with pagination
func (r *VolumeRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Volume{}).Where("deleted_at IS NULL")

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get volumes with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&volumes).Error

	if err != nil {
		return nil, 0, err
	}

	return volumes, total, nil
}

// ListAccessibleByUser lists volumes accessible by user (own volumes + org volumes)
func (r *VolumeRepository) ListAccessibleByUser(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Volume{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("identifier ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get volumes with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&volumes).Error

	if err != nil {
		return nil, 0, err
	}

	return volumes, total, nil
}

// ListByLabel lists volumes by label filters
func (r *VolumeRepository) ListByLabel(ctx context.Context, labels map[string]string, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Volume{}).Where("deleted_at IS NULL")

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

	// Get volumes with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&volumes).Error

	if err != nil {
		return nil, 0, err
	}

	return volumes, total, nil
}

// UpdateLabels updates volume labels
func (r *VolumeRepository) UpdateLabels(ctx context.Context, id int64, labels models.ResourceLabels) error {
	return r.db.WithContext(ctx).Model(&models.Volume{}).
		Where("id = ?", id).
		Update("labels", labels).Error
}

// GetVolumeMounts gets all mounts (workspaces) where this volume is attached
func (r *VolumeRepository) GetVolumeMounts(ctx context.Context, volumeID int64) ([]models.VolumeMount, error) {
	var mounts []models.VolumeMount

	// Query workspace_volumes join to get mount information
	rows, err := r.db.WithContext(ctx).
		Table("workspace_volumes").
		Select("workspace_volumes.workspace_id, workspaces.identifier as workspace_name, workspace_volumes.mount_path").
		Joins("LEFT JOIN workspaces ON workspace_volumes.workspace_id = workspaces.id").
		Where("workspace_volumes.volume_id = ?", volumeID).
		Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query volume mounts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var mount models.VolumeMount
		if err := rows.Scan(&mount.WorkspaceID, &mount.WorkspaceName, &mount.MountPath); err != nil {
			return nil, fmt.Errorf("failed to scan mount row: %w", err)
		}
		mounts = append(mounts, mount)
	}

	return mounts, nil
}

// ListByOrganization lists volumes in a specific organization
func (r *VolumeRepository) ListByOrganization(ctx context.Context, orgID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Volume{}).
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

	// Get volumes with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&volumes).Error

	if err != nil {
		return nil, 0, err
	}

	return volumes, total, nil
}

// ListByOrganizationAll lists all volumes in organizations where user is owner/admin
func (r *VolumeRepository) ListByOrganizationAll(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	// Get organizations where user is owner or admin
	query := r.db.WithContext(ctx).Model(&models.Volume{}).
		Joins("INNER JOIN organizations ON volumes.organization_id = organizations.id").
		Joins("INNER JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ? AND organization_members.role IN ? AND volumes.deleted_at IS NULL", userID, []string{"owner", "admin"})

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("volumes.identifier ILIKE ? OR volumes.display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get volumes with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("volumes.%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&volumes).Error

	if err != nil {
		return nil, 0, err
	}

	return volumes, total, nil
}
