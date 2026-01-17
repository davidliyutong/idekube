package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	redisClient "github.com/davidliyutong/idekube-controller/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID    int64           `json:"user_id"`
	Username  string          `json:"username"`
	Role      models.UserRole `json:"role"`
	TokenType string          `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	RedisClient          *redisClient.Client
}

// JWTManager handles JWT token operations
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config *JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

// generateRefreshTokenID generates a unique refresh token ID
func generateRefreshTokenID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateTokenPair generates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(ctx context.Context, user *models.User) (*models.TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessTokenExpiresAt := now.Add(m.config.AccessTokenDuration)
	accessClaims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "idekube-controller",
			Subject:   user.Username,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token with unique ID
	refreshTokenID, err := generateRefreshTokenID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token ID: %w", err)
	}

	refreshTokenExpiresAt := now.Add(m.config.RefreshTokenDuration)
	refreshClaims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenID,
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "idekube-controller",
			Subject:   user.Username,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store refresh token in Redis
	redisKey := fmt.Sprintf("refresh_token:%d:%s", user.ID, refreshTokenID)
	if err := m.config.RedisClient.Set(ctx, redisKey, refreshTokenString, m.config.RefreshTokenDuration); err != nil {
		return nil, fmt.Errorf("failed to store refresh token in Redis: %w", err)
	}

	return &models.TokenPair{
		AccessToken:           accessTokenString,
		RefreshToken:          refreshTokenString,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		TokenType:             "Bearer",
	}, nil
}

// GenerateToken generates a new JWT token (legacy method for backward compatibility)
func (m *JWTManager) GenerateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.config.AccessTokenDuration)

	claims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "idekube-controller",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (m *JWTManager) RefreshAccessToken(ctx context.Context, refreshTokenString string) (*models.TokenPair, error) {
	// Validate refresh token
	claims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Ensure it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Check if refresh token exists in Redis
	redisKey := fmt.Sprintf("refresh_token:%d:%s", claims.UserID, claims.ID)
	exists, err := m.config.RedisClient.Exists(ctx, redisKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check refresh token in Redis: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// Generate new token pair
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}

	// FIXME: Disabled revocation of old refresh token for now, as it causes issues with multiple concurrent refreshes
	// Revoke old refresh token
	// if err := m.RevokeRefreshToken(ctx, claims.UserID, claims.ID); err != nil {
	// 	// Log error but don't fail the operation
	// 	fmt.Printf("Warning: failed to revoke old refresh token: %v\n", err)
	// }

	return m.GenerateTokenPair(ctx, user)
}

// RevokeRefreshToken revokes a refresh token
func (m *JWTManager) RevokeRefreshToken(ctx context.Context, userID int64, tokenID string) error {
	redisKey := fmt.Sprintf("refresh_token:%d:%s", userID, tokenID)
	return m.config.RedisClient.Del(ctx, redisKey)
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user
func (m *JWTManager) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	// Use Redis SCAN to find all tokens for the user
	pattern := fmt.Sprintf("refresh_token:%d:*", userID)

	// Get the underlying Redis client for SCAN operation
	rdb := m.config.RedisClient.GetClient()
	iter := rdb.Scan(ctx, 0, pattern, 0).Iterator()

	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan refresh tokens: %w", err)
	}

	if len(keys) > 0 {
		return m.config.RedisClient.Del(ctx, keys...)
	}

	return nil
}

// addRandomDelay adds a random delay between 100-500ms to prevent timing attacks
func addRandomDelay() {
	// Random delay between 100 and 500 milliseconds
	delay := time.Duration(100+mathrand.Intn(400)) * time.Millisecond
	time.Sleep(delay)
}

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			addRandomDelay() // Add delay for failed auth
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Authorization header required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			addRandomDelay() // Add delay for failed auth
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			addRandomDelay() // Add delay for failed auth
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Invalid or expired token",
					Details: err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Ensure it's an access token
		if claims.TokenType != "access" {
			addRandomDelay()
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Invalid token type",
				},
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// RequireAuth is a middleware that ensures user authentication
// It should be used after AuthMiddleware to ensure user context is available
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists || userID == nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists || userRole == nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "User role not found",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// MustGetUserID gets the user ID from context
// This function assumes RequireAuth middleware has already validated authentication
// Panics if user_id is not found (should never happen if RequireAuth is used)
func MustGetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		// This should never happen if RequireAuth middleware is used
		panic("user_id not found in context - ensure RequireAuth middleware is applied")
	}
	return userID.(int64)
}

// GetUsername gets the username from context
func GetUsername(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("username not found in context")
	}
	return username.(string), nil
}

func MustGetUserRole(c *gin.Context) models.UserRole {
	role, exists := c.Get("user_role")
	if !exists {
		// This should never happen if RequireAuth middleware is used
		panic("user_role not found in context - ensure RequireAuth middleware is applied")
	}
	return role.(models.UserRole)
}
