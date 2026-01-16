package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TemplateRepository handles template data access
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create creates a new template
func (r *TemplateRepository) Create(ctx context.Context, template *models.Template) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID retrieves a template by ID
func (r *TemplateRepository) GetByID(ctx context.Context, id int64) (*models.Template, error) {
	var template models.Template
	err := r.db.WithContext(ctx).First(&template, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("template not found")
	}
	return &template, err
}

// GetByUUID retrieves a template by UUID
func (r *TemplateRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Template, error) {
	var template models.Template
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&template).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("template not found")
	}
	return &template, err
}

// Update updates a template
func (r *TemplateRepository) Update(ctx context.Context, template *models.Template) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete deletes a template
func (r *TemplateRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Template{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("template not found")
	}
	return nil
}

// ListPublic lists all public templates
func (r *TemplateRepository) ListPublic(ctx context.Context) ([]*models.Template, error) {
	var templates []*models.Template
	err := r.db.WithContext(ctx).Where("is_public = ?", true).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// ListByOwner lists templates owned by a specific owner
func (r *TemplateRepository) ListByOwner(ctx context.Context, ownerType models.OwnerType, ownerID int64) ([]*models.Template, error) {
	var templates []*models.Template
	err := r.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
		Order("created_at DESC").
		Find(&templates).Error
	return templates, err
}

// ListAccessible lists all templates accessible to a user (public + owned)
func (r *TemplateRepository) ListAccessible(ctx context.Context, userID int64, orgIDs []int64) ([]*models.Template, error) {
	var templates []*models.Template

	query := r.db.WithContext(ctx).Where("is_public = ?", true).
		Or("owner_type = ? AND owner_id = ?", "user", userID)

	if len(orgIDs) > 0 {
		query = query.Or("owner_type = ? AND owner_id IN ?", "organization", orgIDs)
	}

	err := query.Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// ListAll retrieves all templates (for admin) with pagination
func (r *TemplateRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.Template, int64, error) {
	var templates []*models.Template
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Template{})

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get templates with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&templates).Error

	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// ListAccessibleByUser lists templates accessible to a user with pagination
func (r *TemplateRepository) ListAccessibleByUser(ctx context.Context, userID int64, orgIDs []int64, opts *models.ListOptions) ([]*models.Template, int64, error) {
	var templates []*models.Template
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Template{}).
		Where("visibility = ?", models.TemplateVisibilityPublic).
		Or("(owner_type = ? AND owner_id = ?)", models.OwnerTypeUser, userID)

	if len(orgIDs) > 0 {
		query = query.Or("(visibility = ? AND owner_type = ? AND owner_id IN ?)",
			models.TemplateVisibilityOrganization, models.OwnerTypeOrganization, orgIDs)
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

	// Get templates with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&templates).Error

	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// ListByVisibility lists templates by visibility level with pagination
func (r *TemplateRepository) ListByVisibility(ctx context.Context, visibility models.TemplateVisibility, opts *models.ListOptions) ([]*models.Template, int64, error) {
	var templates []*models.Template
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Template{}).
		Where("visibility = ?", visibility)

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get templates with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&templates).Error

	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}
