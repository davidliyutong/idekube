package repository

import (
	"context"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// AuditLogRepository handles audit log data access
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List retrieves a list of audit logs with pagination
func (r *AuditLogRepository) List(ctx context.Context, opts *models.ListOptions) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64
	
	// Count total
	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get logs
	offset := (opts.Page - 1) * opts.PageSize
	err := r.db.WithContext(ctx).Order("created_at DESC").
		Limit(opts.PageSize).
		Offset(offset).
		Find(&logs).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// ListByUser retrieves audit logs for a specific user
func (r *AuditLogRepository) ListByUser(ctx context.Context, userID int64, opts *models.ListOptions) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64
	
	// Count total
	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get logs
	offset := (opts.Page - 1) * opts.PageSize
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(opts.PageSize).
		Offset(offset).
		Find(&logs).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// ListByResource retrieves audit logs for a specific resource
func (r *AuditLogRepository) ListByResource(ctx context.Context, resourceType, resourceID string, opts *models.ListOptions) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64
	
	// Count total
	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get logs
	offset := (opts.Page - 1) * opts.PageSize
	err := r.db.WithContext(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Order("created_at DESC").
		Limit(opts.PageSize).
		Offset(offset).
		Find(&logs).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}
