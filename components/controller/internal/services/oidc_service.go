package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"golang.org/x/oauth2"
)

// OIDCService handles OIDC authentication
type OIDCService struct {
	oidcProviderRepo *repository.OIDCProviderRepository
	userRepo         *repository.UserRepository
	sessionRepo      *repository.OAuthSessionRepository
}

// NewOIDCService creates a new OIDC service
func NewOIDCService(
	oidcProviderRepo *repository.OIDCProviderRepository,
	userRepo *repository.UserRepository,
	sessionRepo *repository.OAuthSessionRepository,
) *OIDCService {
	return &OIDCService{
		oidcProviderRepo: oidcProviderRepo,
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
	}
}

// OIDCProvider wraps the provider configuration with oauth2 config
type OIDCProvider struct {
	Provider     *models.OIDCProvider
	Verifier     *oidc.IDTokenVerifier
	OAuth2Config *oauth2.Config
}

// GetProvider retrieves and initializes an OIDC provider
func (s *OIDCService) GetProvider(ctx context.Context, providerName string) (*OIDCProvider, error) {
	// Get provider from database
	provider, err := s.oidcProviderRepo.GetByName(ctx, providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	if !provider.Enabled {
		return nil, fmt.Errorf("provider is disabled")
	}

	// Initialize OIDC provider
	oidcProvider, err := oidc.NewProvider(ctx, provider.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Create verifier
	verifier := oidcProvider.Verifier(&oidc.Config{
		ClientID: provider.ClientID,
	})

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  provider.RedirectURL,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OIDCProvider{
		Provider:     provider,
		Verifier:     verifier,
		OAuth2Config: oauth2Config,
	}, nil
}

// GenerateAuthURL generates the OAuth2 authorization URL
func (s *OIDCService) GenerateAuthURL(ctx context.Context, providerName string) (string, string, error) {
	oidcProvider, err := s.GetProvider(ctx, providerName)
	if err != nil {
		return "", "", err
	}

	// Generate random state
	state, err := generateRandomString(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state in session (you might want to use Redis or similar)
	session := &models.OAuthSession{
		Key:       fmt.Sprintf("oidc_state_%s", state),
		Value:     providerName,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return "", "", fmt.Errorf("failed to store state: %w", err)
	}

	authURL := oidcProvider.OAuth2Config.AuthCodeURL(state)
	return authURL, state, nil
}

// HandleCallback handles the OAuth2 callback
func (s *OIDCService) HandleCallback(ctx context.Context, state, code string) (*models.User, string, error) {
	// Verify state
	session, err := s.sessionRepo.GetByKey(ctx, fmt.Sprintf("oidc_state_%s", state))
	if err != nil {
		return nil, "", fmt.Errorf("invalid state: %w", err)
	}

	providerName := session.Value

	// Delete used state
	_ = s.sessionRepo.DeleteByKey(ctx, fmt.Sprintf("oidc_state_%s", state))

	// Get provider
	oidcProvider, err := s.GetProvider(ctx, providerName)
	if err != nil {
		return nil, "", err
	}

	// Exchange code for token
	oauth2Token, err := oidcProvider.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange token: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, "", fmt.Errorf("no id_token in token response")
	}

	// Verify ID token
	idToken, err := oidcProvider.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Sub           string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, "", fmt.Errorf("failed to parse claims: %w", err)
	}

	// Find or create user
	user, err := s.userRepo.GetByEmail(ctx, claims.Email)
	if err != nil {
		// Create new user
		username := claims.Email
		if len(username) > 50 {
			username = username[:50]
		}

		emailPtr := &claims.Email
		displayName := claims.Name
		user = &models.User{
			Username:      username,
			Email:         emailPtr,
			EmailVerified: claims.EmailVerified,
			DisplayName:   &displayName,
			Role:          models.UserRoleUser,
			Status:        models.UserStatusActive,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user
		if claims.Name != "" && user.DisplayName == nil {
			displayName := claims.Name
			user.DisplayName = &displayName
		}
		if claims.EmailVerified && !user.EmailVerified {
			user.EmailVerified = true
		}
		_ = s.userRepo.Update(ctx, user)
	}

	return user, providerName, nil
}

// CreateProvider creates a new OIDC provider
func (s *OIDCService) CreateProvider(ctx context.Context, req *models.CreateOIDCProviderRequest) (*models.OIDCProvider, error) {
	provider := &models.OIDCProvider{
		Name:         req.Name,
		IssuerURL:    req.IssuerURL,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		RedirectURL:  req.RedirectURL,
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.oidcProviderRepo.Create(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// ListProviders lists all OIDC providers
func (s *OIDCService) ListProviders(ctx context.Context, enabledOnly bool) ([]*models.OIDCProvider, error) {
	if enabledOnly {
		return s.oidcProviderRepo.ListEnabled(ctx)
	}
	return s.oidcProviderRepo.List(ctx)
}

// UpdateProvider updates an OIDC provider
func (s *OIDCService) UpdateProvider(ctx context.Context, id int64, req *models.UpdateOIDCProviderRequest) (*models.OIDCProvider, error) {
	provider, err := s.oidcProviderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.IssuerURL != nil {
		provider.IssuerURL = *req.IssuerURL
	}
	if req.ClientID != nil {
		provider.ClientID = *req.ClientID
	}
	if req.ClientSecret != nil {
		provider.ClientSecret = *req.ClientSecret
	}
	if req.RedirectURL != nil {
		provider.RedirectURL = *req.RedirectURL
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}

	err = s.oidcProviderRepo.Update(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	return provider, nil
}

// DeleteProvider deletes an OIDC provider
func (s *OIDCService) DeleteProvider(ctx context.Context, id int64) error {
	return s.oidcProviderRepo.Delete(ctx, id)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
