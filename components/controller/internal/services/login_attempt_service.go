package services

import (
	"context"
	"fmt"
	"time"

	redisClient "github.com/davidliyutong/idekube-controller/pkg/redis"
)

// LoginAttemptService handles login attempt tracking and banning
type LoginAttemptService struct {
	redis          *redisClient.Client
	settingService *SettingService
}

// NewLoginAttemptService creates a new login attempt service
func NewLoginAttemptService(redis *redisClient.Client, settingService *SettingService) *LoginAttemptService {
	return &LoginAttemptService{
		redis:          redis,
		settingService: settingService,
	}
}

// RecordFailedAttempt records a failed login attempt
func (s *LoginAttemptService) RecordFailedAttempt(ctx context.Context, username string) error {
	key := fmt.Sprintf("login_attempts:%s", username)

	// Get settings dynamically
	maxAttempts := s.settingService.GetSettingAsInt(ctx, "login_max_attempts", 5)
	banDuration := time.Duration(s.settingService.GetSettingAsInt(ctx, "login_ban_duration_minutes", 15)) * time.Minute

	// Increment attempt count
	count, err := s.redis.Incr(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to increment login attempts: %w", err)
	}

	// Set expiration for the first attempt
	if count == 1 {
		if err := s.redis.Expire(ctx, key, banDuration); err != nil {
			return fmt.Errorf("failed to set expiration: %w", err)
		}
	}

	// If max attempts reached, ban the user
	if count >= int64(maxAttempts) {
		banKey := fmt.Sprintf("login_banned:%s", username)
		if err := s.redis.Set(ctx, banKey, "1", banDuration); err != nil {
			return fmt.Errorf("failed to ban user: %w", err)
		}
	}

	return nil
}

// ResetFailedAttempts resets failed login attempts for a user
func (s *LoginAttemptService) ResetFailedAttempts(ctx context.Context, username string) error {
	key := fmt.Sprintf("login_attempts:%s", username)
	banKey := fmt.Sprintf("login_banned:%s", username)

	if err := s.redis.Del(ctx, key, banKey); err != nil {
		return fmt.Errorf("failed to reset login attempts: %w", err)
	}

	return nil
}

// IsUserBanned checks if a user is currently banned
func (s *LoginAttemptService) IsUserBanned(ctx context.Context, username string) (bool, time.Duration, error) {
	banKey := fmt.Sprintf("login_banned:%s", username)

	exists, err := s.redis.Exists(ctx, banKey)
	if err != nil {
		return false, 0, fmt.Errorf("failed to check ban status: %w", err)
	}

	if !exists {
		return false, 0, nil
	}

	// Get remaining ban time
	ttl, err := s.redis.TTL(ctx, banKey)
	if err != nil {
		return true, 0, fmt.Errorf("failed to get ban TTL: %w", err)
	}

	return true, ttl, nil
}

// GetFailedAttempts returns the number of failed attempts for a user
func (s *LoginAttemptService) GetFailedAttempts(ctx context.Context, username string) (int, error) {
	key := fmt.Sprintf("login_attempts:%s", username)

	value, err := s.redis.Get(ctx, key)
	if err != nil {
		// If key doesn't exist, return 0
		return 0, nil
	}

	var count int
	if _, err := fmt.Sscanf(value, "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse attempt count: %w", err)
	}

	return count, nil
}
