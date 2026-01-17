package models

import (
	"time"

	"gorm.io/datatypes"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int64                  `json:"id" db:"id"`
	UserID       *int64                 `json:"user_id,omitempty" db:"user_id"`
	Username     *string                `json:"username,omitempty" db:"username"`
	Action       string                 `json:"action" db:"action"`
	ResourceType *string                `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty" db:"resource_id"`
	Details      datatypes.JSONMap      `json:"details,omitempty" db:"details"`
	IPAddress    *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// CreateAuditLogRequest represents the request to create an audit log
type CreateAuditLogRequest struct {
	UserID       *int64                 `json:"user_id,omitempty"`
	Username     *string                `json:"username,omitempty"`
	Action       string                 `json:"action" binding:"required"`
	ResourceType *string                `json:"resource_type,omitempty"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Details      datatypes.JSONMap `json:"details,omitempty"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
}

// Session represents a user session
type Session struct {
	ID             int64     `json:"id" db:"id"`
	SessionToken   string    `json:"session_token" db:"session_token"`
	UserID         int64     `json:"user_id" db:"user_id"`
	IPAddress      *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string   `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`
}
