package repository

import (
	"context"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// OIDCProviderRepository handles database operations for OIDC providers
type OIDCProviderRepository struct {
	db *gorm.DB
}

// NewOIDCProviderRepository creates a new OIDC provider repository
func NewOIDCProviderRepository(db *gorm.DB) *OIDCProviderRepository {
	return &OIDCProviderRepository{db: db}
}

// Create creates a new OIDC provider
func (r *OIDCProviderRepository) Create(ctx context.Context, provider *models.OIDCProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

// GetByID retrieves an OIDC provider by ID
func (r *OIDCProviderRepository) GetByID(ctx context.Context, id int64) (*models.OIDCProvider, error) {
	var provider models.OIDCProvider
	err := r.db.WithContext(ctx).First(&provider, id).Error
	return &provider, err
}

// GetByName retrieves an OIDC provider by name
func (r *OIDCProviderRepository) GetByName(ctx context.Context, name string) (*models.OIDCProvider, error) {
	var provider models.OIDCProvider
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&provider).Error
	return &provider, err
}

// List retrieves all OIDC providers
func (r *OIDCProviderRepository) List(ctx context.Context) ([]*models.OIDCProvider, error) {
	var providers []*models.OIDCProvider
	err := r.db.WithContext(ctx).Find(&providers).Error
	return providers, err
}

// ListEnabled retrieves all enabled OIDC providers
func (r *OIDCProviderRepository) ListEnabled(ctx context.Context) ([]*models.OIDCProvider, error) {
	var providers []*models.OIDCProvider
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&providers).Error
	return providers, err
}

// Update updates an existing OIDC provider
func (r *OIDCProviderRepository) Update(ctx context.Context, provider *models.OIDCProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// Delete deletes an OIDC provider
func (r *OIDCProviderRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.OIDCProvider{}, id).Error
}
