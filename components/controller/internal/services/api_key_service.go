package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// APIKeyService handles API key operations
type APIKeyService struct {
	apiKeyRepo *repository.APIKeyRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(apiKeyRepo *repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

// CreateAPIKey creates a new API key
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID int64, name string, scopes []string, expiresAt *time.Time) (*models.APIKey, string, error) {
	// Generate random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate key: %w", err)
	}
	plainKey := fmt.Sprintf("idk_%s", base64.URLEncoding.EncodeToString(keyBytes))

	// Hash the key for storage
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash key: %w", err)
	}

	// Create API key record
	apiKey := &models.APIKey{
		UserID:     userID,
		Name:       name,
		KeyHash:    string(hashedKey),
		Scopes:     scopes,
		ExpiresAt:  expiresAt,
		LastUsedAt: nil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.apiKeyRepo.Create(ctx, apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	// Return the plain key (only time it's available)
	return apiKey, plainKey, nil
}

// ValidateAPIKey validates an API key and returns the associated user ID
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, plainKey string) (int64, error) {
	// List all API keys (in production, use a more efficient lookup)
	apiKeys, err := s.apiKeyRepo.List(ctx, 0, 1000)
	if err != nil {
		return 0, fmt.Errorf("failed to list API keys: %w", err)
	}

	for _, apiKey := range apiKeys {
		// Check if key matches
		err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(plainKey))
		if err != nil {
			continue
		}

		// Check if expired
		if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
			return 0, fmt.Errorf("API key expired")
		}

		// Update last used timestamp
		now := time.Now()
		apiKey.LastUsedAt = &now
		_ = s.apiKeyRepo.Update(ctx, apiKey)

		return apiKey.UserID, nil
	}

	return 0, fmt.Errorf("invalid API key")
}

// GetAPIKey retrieves an API key by ID
func (s *APIKeyService) GetAPIKey(ctx context.Context, id, userID int64) (*models.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if apiKey.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	return apiKey, nil
}

// ListAPIKeys lists all API keys for a user
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID int64) ([]*models.APIKey, error) {
	return s.apiKeyRepo.ListByUser(ctx, userID)
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id, userID int64) error {
	apiKey, err := s.GetAPIKey(ctx, id, userID)
	if err != nil {
		return err
	}

	return s.apiKeyRepo.Delete(ctx, apiKey.ID)
}

// UpdateAPIKey updates an API key
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, id, userID int64, updates map[string]interface{}) error {
	apiKey, err := s.GetAPIKey(ctx, id, userID)
	if err != nil {
		return err
	}

	apiKey.UpdatedAt = time.Now()
	return s.apiKeyRepo.Update(ctx, apiKey)
}
