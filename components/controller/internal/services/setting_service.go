package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/redis"
	"go.uber.org/zap"
)

const (
	settingCacheKeyPrefix = "setting:"
	settingCacheTTL       = 5 * time.Minute
	allSettingsCacheKey   = "settings:all"
)

// SettingService handles setting business logic
type SettingService struct {
	settingRepo *repository.SettingRepository
	redisClient *redis.Client
}

// NewSettingService creates a new setting service
func NewSettingService(settingRepo *repository.SettingRepository, redisClient *redis.Client) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		redisClient: redisClient,
	}
}

// GetSetting retrieves a setting by key with caching
func (s *SettingService) GetSetting(ctx context.Context, key string) (*models.Setting, error) {
	// Try to get from cache first
	if s.redisClient != nil {
		cacheKey := settingCacheKeyPrefix + key
		cachedData, err := s.redisClient.Get(ctx, cacheKey)
		if err == nil && cachedData != "" {
			var setting models.Setting
			if err := json.Unmarshal([]byte(cachedData), &setting); err == nil {
				return &setting, nil
			}
		}
	}

	// Get from database
	setting, err := s.settingRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.redisClient != nil {
		cacheKey := settingCacheKeyPrefix + key
		data, err := json.Marshal(setting)
		if err == nil {
			if err := s.redisClient.Set(ctx, cacheKey, string(data), settingCacheTTL); err != nil {
				zap.L().Warn("Failed to cache setting", zap.String("key", key), zap.Error(err))
			}
		}
	}

	return setting, nil
}

// GetAllSettings retrieves all settings with caching
func (s *SettingService) GetAllSettings(ctx context.Context) ([]models.Setting, error) {
	// Try to get from cache first
	if s.redisClient != nil {
		cachedData, err := s.redisClient.Get(ctx, allSettingsCacheKey)
		if err == nil && cachedData != "" {
			var settings []models.Setting
			if err := json.Unmarshal([]byte(cachedData), &settings); err == nil {
				return settings, nil
			}
		}
	}

	// Get from database
	settings, err := s.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.redisClient != nil {
		data, err := json.Marshal(settings)
		if err == nil {
			if err := s.redisClient.Set(ctx, allSettingsCacheKey, string(data), settingCacheTTL); err != nil {
				zap.L().Warn("Failed to cache all settings", zap.Error(err))
			}
		}
	}

	return settings, nil
}

// GetPublicSettings retrieves all public settings (no caching for simplicity)
func (s *SettingService) GetPublicSettings(ctx context.Context) ([]models.Setting, error) {
	return s.settingRepo.GetPublicSettings(ctx)
}

// UpdateSetting updates a setting value and invalidates cache
func (s *SettingService) UpdateSetting(ctx context.Context, key, value string) error {
	// Validate that the setting exists and get its type for validation
	setting, err := s.settingRepo.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	// Validate the value based on type
	if err := s.validateSettingValue(value, setting.ValueType); err != nil {
		return fmt.Errorf("invalid value for setting %s: %w", key, err)
	}

	// Update in database
	if err := s.settingRepo.Update(ctx, key, value); err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(ctx, key)

	return nil
}

// BatchUpdateSettings updates multiple settings and invalidates cache
func (s *SettingService) BatchUpdateSettings(ctx context.Context, updates map[string]string) error {
	// Validate all settings first
	for key, value := range updates {
		setting, err := s.settingRepo.GetByKey(ctx, key)
		if err != nil {
			return fmt.Errorf("setting %s not found: %w", key, err)
		}
		if err := s.validateSettingValue(value, setting.ValueType); err != nil {
			return fmt.Errorf("invalid value for setting %s: %w", key, err)
		}
	}

	// Update in database
	if err := s.settingRepo.BatchUpdate(ctx, updates); err != nil {
		return err
	}

	// Invalidate all caches
	for key := range updates {
		s.invalidateCache(ctx, key)
	}

	return nil
}

// GetSettingAsInt retrieves a setting as integer
func (s *SettingService) GetSettingAsInt(ctx context.Context, key string, defaultValue int) int {
	setting, err := s.GetSetting(ctx, key)
	if err != nil {
		zap.L().Warn("Failed to get setting, using default", zap.String("key", key), zap.Error(err))
		return defaultValue
	}

	value, err := strconv.Atoi(setting.Value)
	if err != nil {
		zap.L().Warn("Failed to parse setting as int, using default", zap.String("key", key), zap.Error(err))
		return defaultValue
	}

	return value
}

// GetSettingAsBool retrieves a setting as boolean
func (s *SettingService) GetSettingAsBool(ctx context.Context, key string, defaultValue bool) bool {
	setting, err := s.GetSetting(ctx, key)
	if err != nil {
		zap.L().Warn("Failed to get setting, using default", zap.String("key", key), zap.Error(err))
		return defaultValue
	}

	value, err := strconv.ParseBool(setting.Value)
	if err != nil {
		zap.L().Warn("Failed to parse setting as bool, using default", zap.String("key", key), zap.Error(err))
		return defaultValue
	}

	return value
}

// GetSettingAsString retrieves a setting as string
func (s *SettingService) GetSettingAsString(ctx context.Context, key string, defaultValue string) string {
	setting, err := s.GetSetting(ctx, key)
	if err != nil {
		zap.L().Warn("Failed to get setting, using default", zap.String("key", key), zap.Error(err))
		return defaultValue
	}

	return setting.Value
}

// validateSettingValue validates a setting value based on its type
func (s *SettingService) validateSettingValue(value string, valueType models.SettingValueType) error {
	switch valueType {
	case models.SettingValueTypeInt:
		_, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("value must be an integer")
		}
	case models.SettingValueTypeBool:
		_, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("value must be a boolean (true/false)")
		}
	case models.SettingValueTypeString:
		// String is always valid
	default:
		return fmt.Errorf("unknown value type: %s", valueType)
	}
	return nil
}

// invalidateCache removes a setting from cache
func (s *SettingService) invalidateCache(ctx context.Context, key string) {
	if s.redisClient == nil {
		return
	}

	// Invalidate specific setting cache
	cacheKey := settingCacheKeyPrefix + key
	if err := s.redisClient.Del(ctx, cacheKey); err != nil {
		zap.L().Warn("Failed to invalidate setting cache", zap.String("key", key), zap.Error(err))
	}

	// Invalidate all settings cache
	if err := s.redisClient.Del(ctx, allSettingsCacheKey); err != nil {
		zap.L().Warn("Failed to invalidate all settings cache", zap.Error(err))
	}
}

// InitializeDefaultSettings ensures default settings exist in database
func (s *SettingService) InitializeDefaultSettings(ctx context.Context) error {
	defaultSettings := []models.Setting{
		{
			Key:         "access_token_expiration_minutes",
			Value:       "15",
			ValueType:   models.SettingValueTypeInt,
			Description: stringPtr("Access token expiration time in minutes"),
			Category:    models.SettingCategoryAuth,
			IsPublic:    false,
		},
		{
			Key:         "refresh_token_expiration_days",
			Value:       "30",
			ValueType:   models.SettingValueTypeInt,
			Description: stringPtr("Refresh token expiration time in days"),
			Category:    models.SettingCategoryAuth,
			IsPublic:    false,
		},
		{
			Key:         "login_max_attempts",
			Value:       "5",
			ValueType:   models.SettingValueTypeInt,
			Description: stringPtr("Maximum login attempts before account is temporarily locked"),
			Category:    models.SettingCategorySecurity,
			IsPublic:    false,
		},
		{
			Key:         "login_ban_duration_minutes",
			Value:       "15",
			ValueType:   models.SettingValueTypeInt,
			Description: stringPtr("Duration of temporary account lock in minutes"),
			Category:    models.SettingCategorySecurity,
			IsPublic:    false,
		},
	}

	for _, setting := range defaultSettings {
		exists, err := s.settingRepo.Exists(ctx, setting.Key)
		if err != nil {
			return fmt.Errorf("failed to check if setting exists: %w", err)
		}

		if !exists {
			if err := s.settingRepo.Create(ctx, &setting); err != nil {
				zap.L().Warn("Failed to create default setting", zap.String("key", setting.Key), zap.Error(err))
				// Continue with other settings
			} else {
				zap.L().Info("Created default setting", zap.String("key", setting.Key))
			}
		}
	}

	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
