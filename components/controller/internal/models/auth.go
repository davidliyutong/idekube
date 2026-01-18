package models

import "time"

// OIDCProvider represents an OIDC provider configuration
type OIDCProvider struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"uniqueIndex;not null"`
	IssuerURL    string    `json:"issuer_url" gorm:"not null"`
	ClientID     string    `json:"client_id" gorm:"not null"`
	ClientSecret string    `json:"-" gorm:"not null"` // Hidden from JSON
	RedirectURL  string    `json:"redirect_url" gorm:"not null"`
	Scopes       []string  `json:"scopes" gorm:"type:text[]"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name
func (OIDCProvider) TableName() string {
	return "oidc_providers"
}

// CreateOIDCProviderRequest represents request to create OIDC provider
type CreateOIDCProviderRequest struct {
	Name         string   `json:"name" binding:"required"`
	IssuerURL    string   `json:"issuer_url" binding:"required"`
	ClientID     string   `json:"client_id" binding:"required"`
	ClientSecret string   `json:"client_secret" binding:"required"`
	RedirectURL  string   `json:"redirect_url" binding:"required"`
	Scopes       []string `json:"scopes"`
}

// UpdateOIDCProviderRequest represents request to update OIDC provider
type UpdateOIDCProviderRequest struct {
	IssuerURL    *string  `json:"issuer_url"`
	ClientID     *string  `json:"client_id"`
	ClientSecret *string  `json:"client_secret"`
	RedirectURL  *string  `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	Enabled      *bool    `json:"enabled"`
}

// OAuthSession represents a temporary session (for OAuth state, tokens, etc.)
type OAuthSession struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name
func (OAuthSession) TableName() string {
	return "oauth_sessions"
}

// MFASetup represents MFA setup response
type MFASetup struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username    string  `json:"username" binding:"required,min=3,max=50"`
	Password    string  `json:"password" binding:"required,min=8"`
	DisplayName string  `json:"display_name" binding:"required,min=1,max=100"`
	Email       *string `json:"email" binding:"omitempty,email"`
}

// RequestPasswordResetRequest represents a password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a password reset with token
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
