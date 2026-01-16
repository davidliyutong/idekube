package repository

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles user data access
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID (excludes soft deleted)
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUUID retrieves a user by UUID
func (r *UserRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username (excludes soft deleted)
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft deletes a user by setting deleted_at
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}
	return nil
}

// List retrieves a list of users with pagination (excludes soft deleted)
func (r *UserRepository) List(ctx context.Context, opts *models.ListOptions) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Build query
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}

// ListAll retrieves all users (for admin) with pagination (excludes soft deleted)
func (r *UserRepository) ListAll(ctx context.Context, opts *models.ListOptions) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ? OR display_name ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// SearchByQuery searches users by username, email, or display name (excludes soft deleted)
func (r *UserRepository) SearchByQuery(ctx context.Context, query string, opts *models.ListOptions) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	searchPattern := "%" + query + "%"
	dbQuery := r.db.WithContext(ctx).Model(&models.User{}).
		Where("deleted_at IS NULL").
		Where("username ILIKE ? OR email ILIKE ? OR display_name ILIKE ?",
			searchPattern, searchPattern, searchPattern)

	// Count total
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	offset := (opts.Page - 1) * opts.PageSize
	err := dbQuery.Order(fmt.Sprintf("%s %s", opts.SortBy, opts.SortOrder)).
		Limit(opts.PageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
