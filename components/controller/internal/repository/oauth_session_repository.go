package repository

import (
	"context"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"gorm.io/gorm"
)

// OAuthSessionRepository handles database operations for OAuth sessions
type OAuthSessionRepository struct {
	db *gorm.DB
}

// NewOAuthSessionRepository creates a new OAuth session repository
func NewOAuthSessionRepository(db *gorm.DB) *OAuthSessionRepository {
	return &OAuthSessionRepository{db: db}
}

// Create creates a new session
func (r *OAuthSessionRepository) Create(ctx context.Context, session *models.OAuthSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByKey retrieves a session by key
func (r *OAuthSessionRepository) GetByKey(ctx context.Context, key string) (*models.OAuthSession, error) {
	var session models.OAuthSession
	err := r.db.WithContext(ctx).
		Where("key = ? AND expires_at > ?", key, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteByKey deletes a session by key
func (r *OAuthSessionRepository) DeleteByKey(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.OAuthSession{}).Error
}

// DeleteExpired deletes expired sessions
func (r *OAuthSessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&models.OAuthSession{}).Error
}

// Update updates a session
func (r *OAuthSessionRepository) Update(ctx context.Context, session *models.OAuthSession) error {
	session.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(session).Error
}
