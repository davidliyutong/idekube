package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	"github.com/davidliyutong/idekube-controller/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserService handles user business logic
type UserService struct {
	userRepo            *repository.UserRepository
	jwtManager          *middleware.JWTManager
	eventPublisher      *queue.EventPublisher
	loginAttemptService *LoginAttemptService
}

// NewUserService creates a new user service
func NewUserService(
	userRepo *repository.UserRepository,
	jwtManager *middleware.JWTManager,
	eventPublisher *queue.EventPublisher,
	loginAttemptService *LoginAttemptService,
) *UserService {
	return &UserService{
		userRepo:            userRepo,
		jwtManager:          jwtManager,
		eventPublisher:      eventPublisher,
		loginAttemptService: loginAttemptService,
	}
}

// Login authenticates a user and returns a token
func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Check if user is banned
	if s.loginAttemptService != nil {
		banned, remainingTime, err := s.loginAttemptService.IsUserBanned(ctx, req.Username)
		if err != nil {
			zap.L().Error("Failed to check ban status", zap.Error(err))
		} else if banned {
			return nil, fmt.Errorf("account temporarily locked due to too many failed login attempts, try again in %v", remainingTime.Round(time.Second))
		}
	}

	// Get user by username (using Identifier field)
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Record failed attempt
		if s.loginAttemptService != nil {
			if err := s.loginAttemptService.RecordFailedAttempt(ctx, req.Username); err != nil {
				zap.L().Error("Failed to record failed login attempt", zap.Error(err))
			}
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active (using Base.Status)
	if user.Status != models.UserStatusActive {
		return nil, fmt.Errorf("user account is not active")
	}

	// Verify password (using Security.PasswordHash)
	err = utils.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		// Record failed attempt
		if s.loginAttemptService != nil {
			if err := s.loginAttemptService.RecordFailedAttempt(ctx, req.Username); err != nil {
				zap.L().Error("Failed to record failed login attempt", zap.Error(err))
			}
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed attempts on successful login
	if s.loginAttemptService != nil {
		if err := s.loginAttemptService.ResetFailedAttempts(ctx, req.Username); err != nil {
			zap.L().Error("Failed to reset failed login attempts", zap.Error(err))
		}
	}

	// Generate JWT token pair
	tokenPair, err := s.jwtManager.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &models.LoginResponse{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
		TokenType:             tokenPair.TokenType,
		User:                  user,
	}, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not specified
	role := req.Role
	if role == "" {
		role = models.UserRoleUser
	}

	now := time.Now()
	user := &models.User{
		Base: models.Base{
			UUID:      uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    models.UserStatusActive,
			ExtraInfo: req.ExtraInfo,
		},
		Profile: models.Profile{
			Identifier:  req.Username,
			DisplayName: req.DisplayName,
		},
		Security: models.Security{
			PasswordHash: hashedPassword,
		},
		Email: req.Email,
		Role:  role,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers lists users with pagination
func (s *UserService) ListUsers(ctx context.Context, opts *models.ListOptions) ([]*models.User, int64, error) {
	return s.userRepo.List(ctx, opts)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		user.IconURL = req.IconURL
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.ExtraInfo != nil {
		user.ExtraInfo = req.ExtraInfo
	}
	if req.Labels != nil {
		user.Labels = req.Labels
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user and publishes event for K8S cleanup
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	// Get user info before deletion
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete the user
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Publish delete event to HouseKeeper for K8S resource cleanup
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishUserDelete(ctx, id, user.Identifier); err != nil {
			// Log error but don't fail the operation
			// HouseKeeper reconciler will handle cleanup eventually
			zap.L().Error("Failed to publish user delete event",
				zap.Int64("user_id", id),
				zap.Error(err))
		}
	}

	return nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID int64, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	err = utils.VerifyPassword(req.OldPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hashedPassword

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateUserProfile updates user's own profile (limited fields)
func (s *UserService) UpdateUserProfile(ctx context.Context, userID int64, req *models.UpdateUserProfileRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update allowed fields
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		user.IconURL = req.IconURL
	}
	if req.Description != nil {
		user.Description = req.Description
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

// UpdateUserByAdmin updates user by admin (more fields allowed)
func (s *UserService) UpdateUserByAdmin(ctx context.Context, userID int64, req *models.UpdateUserRequest, isAdmin bool) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields that admin can change
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		user.IconURL = req.IconURL
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	// Only super admin can change roles
	if req.Role != nil {
		if !isAdmin {
			return nil, fmt.Errorf("only super admin can change user roles")
		}
		user.Role = *req.Role
	}

	if req.ExtraInfo != nil {
		user.ExtraInfo = req.ExtraInfo
	}
	if req.Labels != nil {
		user.Labels = req.Labels
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// CheckUserExists checks if a user exists by username (for power_user)
func (s *UserService) CheckUserExists(ctx context.Context, username string) (bool, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if err.Error() == "user not found" {
			return false, nil
		}
		return false, err
	}
	return user != nil, nil
}

// SearchUsers searches users by query (for organization owner to add members)
func (s *UserService) SearchUsers(ctx context.Context, query string, opts *models.ListOptions) ([]*models.User, int64, error) {
	return s.userRepo.SearchByQuery(ctx, query, opts)
}

// ListAllUsers lists all users (for admin)
func (s *UserService) ListAllUsers(ctx context.Context, opts *models.ListOptions) ([]*models.User, int64, error) {
	return s.userRepo.ListAll(ctx, opts)
}

// RefreshToken generates a new access token from a refresh token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Generate new token pair using JWT manager
	tokenPair, err := s.jwtManager.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user info (from the claims embedded in the new tokens)
	// We validate the access token to extract user info
	claims, err := s.jwtManager.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate new access token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &models.LoginResponse{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
		TokenType:             tokenPair.TokenType,
		User:                  user,
	}, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *UserService) RevokeRefreshToken(ctx context.Context, userID int64, tokenID string) error {
	return s.jwtManager.RevokeRefreshToken(ctx, userID, tokenID)
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user
func (s *UserService) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	return s.jwtManager.RevokeAllRefreshTokens(ctx, userID)
}

// ============================================================================
// Sub-resource APIs
// ============================================================================

// GetUserProfile returns the user's profile sub-resource
func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*models.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.UserProfileResponse{
		Identifier:  user.Identifier,
		DisplayName: user.DisplayName,
		IconURL:     user.IconURL,
		Description: user.Description,
	}, nil
}

// UpdateUserProfileSubResource updates the user's profile sub-resource
func (s *UserService) UpdateUserProfileSubResource(ctx context.Context, userID int64, req *models.UpdateUserProfileRequest) (*models.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update profile fields
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.IconURL != nil {
		user.IconURL = req.IconURL
	}
	if req.Description != nil {
		user.Description = req.Description
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &models.UserProfileResponse{
		Identifier:  user.Identifier,
		DisplayName: user.DisplayName,
		IconURL:     user.IconURL,
		Description: user.Description,
	}, nil
}

// GetUserSecurity returns the user's security sub-resource (no sensitive data)
func (s *UserService) GetUserSecurity(ctx context.Context, userID int64) (*models.UserSecurityResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.UserSecurityResponse{
		MFAEnabled:     user.MFAEnabled,
		HasBackupCodes: len(user.MFABackupCodes) > 0,
		LastLoginAt:    user.LastLoginAt,
	}, nil
}

// UpdateUserSecurity updates the user's security sub-resource
func (s *UserService) UpdateUserSecurity(ctx context.Context, userID int64, req *models.UpdateUserSecurityRequest) (*models.UserSecurityResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update password if provided
	if req.OldPassword != nil && req.NewPassword != nil {
		// Verify old password
		err = utils.VerifyPassword(*req.OldPassword, user.PasswordHash)
		if err != nil {
			return nil, fmt.Errorf("invalid old password")
		}

		// Hash new password
		hashedPassword, err := utils.HashPassword(*req.NewPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = hashedPassword
	}

	// Update MFA enabled status if provided
	if req.MFAEnabled != nil {
		user.MFAEnabled = *req.MFAEnabled
		// If disabling MFA, clear the secret and backup codes
		if !*req.MFAEnabled {
			user.MFASecret = nil
			user.MFABackupCodes = nil
		}
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update security settings: %w", err)
	}

	return &models.UserSecurityResponse{
		MFAEnabled:     user.MFAEnabled,
		HasBackupCodes: len(user.MFABackupCodes) > 0,
		LastLoginAt:    user.LastLoginAt,
	}, nil
}

// GetUserEmail returns the user's email address and verification status
func (s *UserService) GetUserEmail(ctx context.Context, userID int64) (*models.UserEmailResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.UserEmailResponse{
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
	}, nil
}

// UpdateUserEmail updates the user's email address and sets IsEmailVerified to false
func (s *UserService) UpdateUserEmail(ctx context.Context, userID int64, req *models.UpdateUserEmailRequest) (*models.UserEmailResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update email and reset verification status
	user.Email = &req.Email
	user.IsEmailVerified = false

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update email: %w", err)
	}

	// TODO: Publish event for email verification when needed

	return &models.UserEmailResponse{
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
	}, nil
}
