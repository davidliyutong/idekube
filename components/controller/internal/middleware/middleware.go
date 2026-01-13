package middleware

import (
	"net/http"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditMiddleware logs all API requests to audit logs
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	auditRepo := repository.NewAuditLogRepository(db)
	
	return func(c *gin.Context) {
		// Process request
		c.Next()
		
		// Skip audit for health checks and metrics
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			return
		}
		
		// Get user info from context (if authenticated)
		var userID *int64
		var username *string
		
		if uid, exists := c.Get("user_id"); exists {
			id := uid.(int64)
			userID = &id
		}
		
		if uname, exists := c.Get("username"); exists {
			name := uname.(string)
			username = &name
		}
		
		// Get IP and user agent
		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()
		
		// Determine action from HTTP method and path
		action := c.Request.Method + " " + c.Request.URL.Path
		
		// Create audit log
		log := &models.AuditLog{
			UserID:    userID,
			Username:  username,
			Action:    action,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
			Details: map[string]interface{}{
				"status_code": c.Writer.Status(),
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"query":       c.Request.URL.RawQuery,
			},
		}
		
		// Save audit log asynchronously (don't block request)
		go func() {
			_ = auditRepo.Create(c.Request.Context(), log)
		}()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// TODO: Implement proper request ID generation
	return "req-" + randomString(16)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
