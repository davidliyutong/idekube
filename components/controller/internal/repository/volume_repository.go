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

// GetByID retrieves a volume by ID
func (r *VolumeRepository) GetByID(ctx context.Context, id int64) (*models.Volume, error) {
	var volume models.Volume
	err := r.db.WithContext(ctx).First(&volume, id).Error
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

// Delete deletes a volume
func (r *VolumeRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Volume{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("volume not found")
	}
	return nil
}

// ListByOwner lists volumes owned by a specific owner
func (r *VolumeRepository) ListByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) ([]*models.Volume, error) {
	var volumes []*models.Volume
	err := r.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
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
