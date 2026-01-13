package repository

import (
	"context"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// WebhookRepository handles database operations for webhooks
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create creates a new webhook
func (r *WebhookRepository) Create(ctx context.Context, webhook *models.Webhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

// GetByID retrieves a webhook by ID
func (r *WebhookRepository) GetByID(ctx context.Context, id int64) (*models.Webhook, error) {
	var webhook models.Webhook
	err := r.db.WithContext(ctx).First(&webhook, id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// ListByUser lists all webhooks for a user
func (r *WebhookRepository) ListByUser(ctx context.Context, userID int64) ([]*models.Webhook, error) {
	var webhooks []*models.Webhook
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&webhooks).Error
	return webhooks, err
}

// ListByUserAndEvents lists webhooks for a user filtered by events
func (r *WebhookRepository) ListByUserAndEvents(ctx context.Context, userID int64, events []string) ([]*models.Webhook, error) {
	var webhooks []*models.Webhook
	query := r.db.WithContext(ctx).Where("user_id = ? AND enabled = ?", userID, true)

	if len(events) > 0 {
		// Filter webhooks that have at least one of the specified events
		query = query.Where("EXISTS (SELECT 1 FROM unnest(events) AS event WHERE event = ANY(?))", events)
	}

	err := query.Order("created_at DESC").Find(&webhooks).Error
	return webhooks, err
}

// Update updates a webhook
func (r *WebhookRepository) Update(ctx context.Context, webhook *models.Webhook, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(webhook).Updates(updates).Error
}

// Delete deletes a webhook
func (r *WebhookRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Webhook{}, id).Error
}
