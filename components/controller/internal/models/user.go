package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UserRole represents user role enum
type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin"
	UserRoleAdmin      UserRole = "admin"
	UserRolePowerUser  UserRole = "power_user"
	UserRoleUser       UserRole = "user"
)

// UserStatus represents user status enum
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// User represents a user in the system
type User struct {
	ID             int64             `json:"id" db:"id"`
	UUID           uuid.UUID         `json:"uuid" db:"uuid"`
	Username       string            `json:"username" db:"username"`
	Email          *string           `json:"email,omitempty" db:"email"`
	EmailVerified  bool              `json:"email_verified" db:"email_verified"`
	PasswordHash   string            `json:"-" db:"password_hash"`
	Role           UserRole          `json:"role" db:"role"`
	Status         UserStatus        `json:"status" db:"status"`
	AvatarURL      *string           `json:"avatar_url,omitempty" db:"avatar_url"`
	DisplayName    *string           `json:"display_name,omitempty" db:"display_name"`
	ExtraInfo      datatypes.JSONMap `json:"extra_info,omitempty" db:"extra_info"`
	MFAEnabled     bool              `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret      *string           `json:"-" db:"mfa_secret"`
	MFABackupCodes []string          `json:"-" db:"mfa_backup_codes" gorm:"type:text[]"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
	LastLoginAt    *time.Time        `json:"last_login_at,omitempty" db:"last_login_at"`
	DeletedAt      *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username    string            `json:"username" binding:"required,min=3,max=255"`
	Email       *string           `json:"email,omitempty" binding:"omitempty,email"`
	Password    string            `json:"password" binding:"required,min=8"`
	Role        UserRole          `json:"role,omitempty"`
	DisplayName *string           `json:"display_name,omitempty"`
	ExtraInfo   datatypes.JSONMap `json:"extra_info,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email       *string           `json:"email,omitempty" binding:"omitempty,email"`
	DisplayName *string           `json:"display_name,omitempty"`
	AvatarURL   *string           `json:"avatar_url,omitempty"`
	Status      *UserStatus       `json:"status,omitempty"`
	Role        *UserRole         `json:"role,omitempty"`
	ExtraInfo   datatypes.JSONMap `json:"extra_info,omitempty"`
}

// ChangePasswordRequest represents the request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
	User                  *User     `json:"user"`
}

// UpdateUserProfileRequest represents the request for a user to update their own profile
type UpdateUserProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty" binding:"omitempty,email"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// CheckUserExistsResponse represents the response for checking if a user exists
type CheckUserExistsResponse struct {
	Exists   bool   `json:"exists"`
	Username string `json:"username"`
}
