package services

import (
	"context"
	"fmt"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/utils"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo            *repository.UserRepository
	jwtManager          *middleware.JWTManager
	loginAttemptService *LoginAttemptService
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	jwtManager *middleware.JWTManager,
	loginAttemptService *LoginAttemptService,
) *AuthService {
	return &AuthService{
		userRepo:            userRepo,
		jwtManager:          jwtManager,
		loginAttemptService: loginAttemptService,
	}
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Check if user is banned
	if s.loginAttemptService != nil {
		banned, remainingTime, err := s.loginAttemptService.IsUserBanned(ctx, req.Username)
		if err == nil && banned {
			return nil, fmt.Errorf("too many failed login attempts, please try again in %v", remainingTime)
		}
	}

	// Get user by username (using Identifier field)
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Record failed attempt
		if s.loginAttemptService != nil {
			_ = s.loginAttemptService.RecordFailedAttempt(ctx, req.Username)
		}
		return nil, fmt.Errorf("invalid username or password")
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
			_ = s.loginAttemptService.RecordFailedAttempt(ctx, req.Username)
		}
		return nil, fmt.Errorf("invalid username or password")
	}

	// Reset failed attempts on successful login
	if s.loginAttemptService != nil {
		_ = s.loginAttemptService.ResetFailedAttempts(ctx, req.Username)
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

// RefreshToken generates a new access token from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Generate new token pair using JWT manager
	tokenPair, err := s.jwtManager.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token: %w", err)
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

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if username already exists
	existingUser, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if req.Email != nil {
		existingUser, _ := s.userRepo.GetByEmail(ctx, *req.Email)
		if existingUser != nil {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	displayName := req.DisplayName
	user := &models.User{
		Base: models.Base{
			UUID:      uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Status:    models.UserStatusActive,
		},
		Profile: models.Profile{
			Identifier:  req.Username,
			DisplayName: &displayName,
		},
		Security: models.Security{
			PasswordHash: hashedPassword,
		},
		Email: req.Email,
		Role:  models.UserRoleUser, // Default role for registration
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Logout revokes user's tokens (placeholder for future implementation)
func (s *AuthService) Logout(ctx context.Context, userID int64, tokenID string) error {
	// TODO: Implement token revocation logic
	// For now, this is a placeholder as JWT tokens are stateless
	// You might want to implement a token blacklist using Redis
	return nil
}
