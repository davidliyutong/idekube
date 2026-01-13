package repository

import (
	"context"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// QuotaRepository handles database operations for quotas
type QuotaRepository struct {
	db *gorm.DB
}

// NewQuotaRepository creates a new quota repository
func NewQuotaRepository(db *gorm.DB) *QuotaRepository {
	return &QuotaRepository{db: db}
}

// GetByOwner retrieves a quota by owner type and ID
func (r *QuotaRepository) GetByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) (*models.Quota, error) {
	var quota models.Quota
	err := r.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).First(&quota).Error
	return &quota, err
}

// Create creates a new quota
func (r *QuotaRepository) Create(ctx context.Context, quota *models.Quota) error {
	return r.db.WithContext(ctx).Create(quota).Error
}

// Update updates an existing quota
func (r *QuotaRepository) Update(ctx context.Context, quota *models.Quota) error {
	return r.db.WithContext(ctx).Save(quota).Error
}

// Delete deletes a quota
func (r *QuotaRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Quota{}, id).Error
}
