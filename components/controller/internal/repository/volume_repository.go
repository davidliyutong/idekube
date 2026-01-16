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

// ListByOwner lists volumes owned by a specific owner (excludes soft deleted)
func (r *VolumeRepository) ListByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) ([]*models.Volume, error) {
	var volumes []*models.Volume
	err := r.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ? AND deleted_at IS NULL", ownerType, ownerID).
		Order("created_at DESC").
		Find(&volumes).Error
	return volumes, err
}

// UpdateStatus updates volume status
func (r *VolumeRepository) UpdateStatus(ctx context.Context, id int64, status models.VolumeStatus, pvcName *string) error {
	return r.db.WithContext(ctx).Model(&models.Volume{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   status,
			"pvc_name": pvcName,
		}).Error
}

// ListAll retrieves all volumes (for admin) with pagination
func (r *VolumeRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Volume{})

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
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

// ListAccessibleByUser lists volumes accessible by user (own volumes only for now)
func (r *VolumeRepository) ListAccessibleByUser(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.Volume, int64, error) {
	var volumes []*models.Volume
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Volume{}).
		Where("owner_type = ? AND owner_id = ?", models.OwnerTypeUser, userID)

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
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

	query := r.db.WithContext(ctx).Model(&models.Volume{})

	// Apply label filters
	for key, value := range labels {
		query = query.Where("labels->>? = ?", key, value)
	}

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
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
