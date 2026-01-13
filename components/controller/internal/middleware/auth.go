package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID   int64            `json:"user_id"`
	Username string           `json:"username"`
	Role     models.UserRole  `json:"role"`
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

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
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

// RequireRole is a middleware that checks if user has required role
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
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
		
		role := userRole.(models.UserRole)
		
		// Super admin has access to everything
		if role == models.UserRoleSuperAdmin {
			c.Next()
			return
		}
		
		// Check if user has required role
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}
		
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
		})
		c.Abort()
	}
}

// GetUserID gets the user ID from context
func GetUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID.(int64), nil
}

// GetUsername gets the username from context
func GetUsername(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("username not found in context")
	}
	return username.(string), nil
}

// GetUserRole gets the user role from context
func GetUserRole(c *gin.Context) (models.UserRole, error) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", fmt.Errorf("user role not found in context")
	}
	return role.(models.UserRole), nil
}
