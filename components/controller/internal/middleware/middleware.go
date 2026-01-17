package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditMiddleware logs all API requests to audit logs
func AuditMiddleware(db *gorm.DB, logger *logger.Logger) gin.HandlerFunc {
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

		// FIXME: ResourceType and ResourceID could be improved by parsing the path
		resourceType := "Unknown"
		resourceID := ""

		// Create audit log
		log := &models.AuditLog{
			UserID:       userID,
			Username:     username,
			ResourceType: &resourceType,
			ResourceID:   &resourceID,
			Action:       action,
			IPAddress:    &ipAddress,
			UserAgent:    &userAgent,
			Details: datatypes.JSONMap{
				"status_code": c.Writer.Status(),
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"query":       c.Request.URL.RawQuery,
			},
		}

		// Save audit log asynchronously (don't block request)
		go func(auditLog *models.AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := auditRepo.Create(ctx, auditLog)
			if err != nil {
				// Log the error (in real implementation, use proper logging)
				logger.Error("Failed to create audit log", "error", zap.Error(err))
			}
		}(log)
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

// SecurityMiddleware adds security headers to all responses
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		// Enable browser XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		// Prevent content sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// Force HTTPS transport (31536000 seconds = 1 year)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Content Security Policy - restricts resource loading
		c.Header("Content-Security-Policy", "default-src 'self'")
		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Permissions policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// visitor tracks request timestamps for rate limiting
type visitor struct {
	lastSeen time.Time
	requests []time.Time
	mu       sync.Mutex
}

// RateLimiter implements IP-based rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	config   RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		config:   config,
	}

	// Start cleanup goroutine to remove old visitors
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes visitors that haven't been seen in 3 minutes
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getVisitor returns or creates a visitor for the given IP
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			lastSeen: time.Now(),
			requests: []time.Time{},
		}
		rl.visitors[ip] = v
	}

	return v
}

// isAllowed checks if a request from the given IP is allowed
func (rl *RateLimiter) isAllowed(ip string) bool {
	v := rl.getVisitor(ip)

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	v.lastSeen = now

	// Remove requests older than 1 minute
	cutoff := now.Add(-1 * time.Minute)
	newRequests := []time.Time{}
	for _, t := range v.requests {
		if t.After(cutoff) {
			newRequests = append(newRequests, t)
		}
	}
	v.requests = newRequests

	// Check if we've exceeded the rate limit
	if len(v.requests) >= rl.config.RequestsPerMinute {
		return false
	}

	// Add current request
	v.requests = append(v.requests, now)
	return true
}

// RateLimitMiddleware limits requests per IP address
// It considers X-Forwarded-For and X-Real-IP headers for reverse proxy support
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get real IP address, considering reverse proxy headers
		ip := getRealIP(c)

		if !limiter.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRealIP extracts the real client IP address
// It checks X-Forwarded-For and X-Real-IP headers first (for reverse proxy support)
func getRealIP(c *gin.Context) string {
	// Check X-Real-IP first
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// Check X-Forwarded-For (take the first IP if multiple)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first one (the client IP)
		for idx := 0; idx < len(xff); idx++ {
			if xff[idx] == ',' {
				return xff[:idx]
			}
		}
		return xff
	}

	// Fall back to remote address
	return c.ClientIP()
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// MaliciousRouteInterceptor handles 404 errors and blocks suspicious patterns
func MaliciousRouteInterceptor() gin.HandlerFunc {
	suspiciousPatterns := []string{
		"/.env",
		"/wp-admin",
		"/phpmyadmin",
		"/.git",
		"/admin",
		"/shell",
		"/config",
		"/.aws",
		"/actuator",
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Check for suspicious patterns
		for _, pattern := range suspiciousPatterns {
			if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
				c.JSON(http.StatusNotFound, models.APIResponse{
					Success: false,
					Error: &models.APIError{
						Code:    "NOT_FOUND",
						Message: "Resource not found",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()

		// Handle 404s for routes that don't match any handler
		if c.Writer.Status() == http.StatusNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "NOT_FOUND",
					Message: "Resource not found",
				},
			})
		}
	}
}
