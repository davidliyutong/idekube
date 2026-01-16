package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// SettingRepository handles setting data access
type SettingRepository struct {
	db *gorm.DB
}

// NewSettingRepository creates a new setting repository
func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

// GetByKey retrieves a setting by key
func (r *SettingRepository) GetByKey(ctx context.Context, key string) (*models.Setting, error) {
	var setting models.Setting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("setting not found")
	}
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetAll retrieves all settings
func (r *SettingRepository) GetAll(ctx context.Context) ([]models.Setting, error) {
	var settings []models.Setting
	err := r.db.WithContext(ctx).Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// GetByCategory retrieves settings by category
func (r *SettingRepository) GetByCategory(ctx context.Context, category models.SettingCategory) ([]models.Setting, error) {
	var settings []models.Setting
	err := r.db.WithContext(ctx).Where("category = ?", category).Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// GetPublicSettings retrieves all public settings
func (r *SettingRepository) GetPublicSettings(ctx context.Context) ([]models.Setting, error) {
	var settings []models.Setting
	err := r.db.WithContext(ctx).Where("is_public = ?", true).Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// Update updates a setting value
func (r *SettingRepository) Update(ctx context.Context, key, value string) error {
	result := r.db.WithContext(ctx).Model(&models.Setting{}).
		Where("key = ?", key).
		Update("value", value)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("setting not found")
	}

	return nil
}

// BatchUpdate updates multiple settings
func (r *SettingRepository) BatchUpdate(ctx context.Context, updates map[string]string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range updates {
			result := tx.Model(&models.Setting{}).
				Where("key = ?", key).
				Update("value", value)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("setting not found: %s", key)
			}
		}
		return nil
	})
}

// Create creates a new setting
func (r *SettingRepository) Create(ctx context.Context, setting *models.Setting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

// Delete deletes a setting by key
func (r *SettingRepository) Delete(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.Setting{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("setting not found")
	}
	return nil
}

// Exists checks if a setting exists
func (r *SettingRepository) Exists(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Setting{}).Where("key = ?", key).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
