package models

import "time"

// Webhook represents a webhook subscription
type Webhook struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	UserID          int64      `json:"user_id" gorm:"not null;index"`
	URL             string     `json:"url" gorm:"not null"`
	Events          []string   `json:"events" gorm:"type:text[];not null"`
	Secret          *string    `json:"-"` // For HMAC signature
	Enabled         bool       `json:"enabled" gorm:"default:true"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName specifies the table name
func (Webhook) TableName() string {
	return "webhooks"
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID         int64      `json:"id" gorm:"primaryKey"`
	UserID     int64      `json:"user_id" gorm:"not null;index"`
	Name       string     `json:"name" gorm:"not null"`
	KeyHash    string     `json:"-" gorm:"not null;uniqueIndex"` // Hashed key
	Scopes     []string   `json:"scopes" gorm:"type:text[]"`
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TableName specifies the table name
func (APIKey) TableName() string {
	return "api_keys"
}
