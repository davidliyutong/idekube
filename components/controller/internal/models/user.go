package models

import (
	"time"

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

// UserStatus constants (for business logic validation, stored as Base.Status string)
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
)

// User represents a user in the system
// Embeds Base (ID, UUID, CreatedAt, UpdatedAt, DeletedAt, Labels, Status, ExtraInfo)
// Embeds Profile (Identifier as username, DisplayName, IconURL, Description)
// Embeds Security (PasswordHash, MFA-related fields, LastLoginAt)
type User struct {
	Base                       // Embedded Base fields
	Profile  `gorm:"embedded"` // Embedded Profile fields (Identifier serves as username)
	Security `gorm:"embedded"` // Embedded Security fields
	Email            *string  `json:"email,omitempty" gorm:"type:varchar(255);uniqueIndex"`
	IsEmailVerified  bool     `json:"is_email_verified" gorm:"column:email_verified;default:false"`
	Role             UserRole `json:"role" gorm:"type:varchar(50);default:'user'"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// GetUsername returns the username (alias for Identifier for backward compatibility)
func (u *User) GetUsername() string {
	return u.Identifier
}

// SetUsername sets the username (alias for Identifier for backward compatibility)
func (u *User) SetUsername(username string) {
	u.Identifier = username
}

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username    string            `json:"username" binding:"required,min=3,max=255"`
	Email       *string           `json:"email,omitempty" binding:"omitempty,email"`
	Password    string            `json:"password" binding:"required,min=8"`
	Role        UserRole          `json:"role,omitempty"`
	DisplayName *string           `json:"display_name,omitempty"`
	ExtraInfo   datatypes.JSONMap `json:"extra_info,omitempty"`
}

// UpdateUserRequest represents the request to update a user (admin operation)
type UpdateUserRequest struct {
	Email       *string           `json:"email,omitempty" binding:"omitempty,email"`
	DisplayName *string           `json:"display_name,omitempty"`
	IconURL     *string           `json:"icon_url,omitempty"`
	Status      *string           `json:"status,omitempty"`
	Role        *UserRole         `json:"role,omitempty"`
	ExtraInfo   datatypes.JSONMap `json:"extra_info,omitempty"`
	Labels      datatypes.JSONMap `json:"labels,omitempty"`
}

// ChangePasswordRequest represents the request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	IPAddress string `json:"-"` // Internal field, not from JSON
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
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateUserSecurityRequest represents the request to update user security settings
type UpdateUserSecurityRequest struct {
	OldPassword *string `json:"old_password,omitempty"`
	NewPassword *string `json:"new_password,omitempty" binding:"omitempty,min=8"`
	MFAEnabled  *bool   `json:"mfa_enabled,omitempty"`
}

// CheckUserExistsResponse represents the response for checking if a user exists
type CheckUserExistsResponse struct {
	Exists   bool   `json:"exists"`
	Username string `json:"username"`
}

// UserProfileResponse represents the user profile sub-resource response
type UserProfileResponse struct {
	Identifier  string  `json:"identifier"`
	DisplayName *string `json:"display_name,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UserSecurityResponse represents the user security sub-resource response (no sensitive data)
type UserSecurityResponse struct {
	MFAEnabled       bool       `json:"mfa_enabled"`
	HasBackupCodes   bool       `json:"has_backup_codes"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	PasswordLastSet  *time.Time `json:"password_last_set,omitempty"`
}

// UserEmailResponse represents the user email sub-resource response
type UserEmailResponse struct {
	Email           *string `json:"email,omitempty"`
	IsEmailVerified bool    `json:"is_email_verified"`
}

// UpdateUserEmailRequest represents the request to update user email
type UpdateUserEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}
