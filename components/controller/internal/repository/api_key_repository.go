package repository

import (
	"context"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// APIKeyRepository handles database operations for API keys
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepository) GetByID(ctx context.Context, id int64) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, id).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// ListByUser lists all API keys for a user
func (r *APIKeyRepository) ListByUser(ctx context.Context, userID int64) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

// List lists API keys with pagination
func (r *APIKeyRepository) List(ctx context.Context, offset, limit int) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

// Update updates an API key
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *models.APIKey) error {
	apiKey.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(apiKey).Error
}

// Delete deletes an API key
func (r *APIKeyRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.APIKey{}, id).Error
}
