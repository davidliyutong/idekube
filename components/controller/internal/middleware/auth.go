package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID   int64           `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

// JWTManager handles JWT token operations
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config *JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

// GenerateToken generates a new JWT token
func (m *JWTManager) GenerateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.config.TokenDuration)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
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

// addRandomDelay adds a random delay between 100-500ms to prevent timing attacks
func addRandomDelay() {
	// Random delay between 100 and 500 milliseconds
	delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
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
