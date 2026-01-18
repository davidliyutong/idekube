package api

import (
	"context"
	"net/http"
	"time"

	_ "github.com/davidliyutong/idekube-controller/docs/api"
	"github.com/davidliyutong/idekube-controller/internal/handlers"
	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/version"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Server represents the API server
type Server struct {
	router      *gin.Engine
	httpServer  *http.Server
	db          *gorm.DB
	jwtManager  *middleware.JWTManager
	rateLimiter *middleware.RateLimiter
	logger      *logger.Logger
}

// NewServer creates a new API server
func NewServer(db *gorm.DB, jwtManager *middleware.JWTManager, logger *logger.Logger) *Server {
	// Create rate limiter: 100 requests per minute per IP
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         20,
	})

	return &Server{
		router:      gin.Default(),
		db:          db,
		jwtManager:  jwtManager,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// SetupRoutes sets up all API routes
func (s *Server) SetupRoutes(
	authHandler *handlers.AuthHandler,
	accountHandler *handlers.AccountHandler,
	userHandler *handlers.UserHandler,
	mfaHandler *handlers.MFAHandler,
	orgHandler *handlers.OrganizationHandler,
	templateHandler *handlers.TemplateHandler,
	workspaceHandler *handlers.WorkspaceHandler,
	volumeHandler *handlers.VolumeHandler,
	settingHandler *handlers.SettingHandler,
	emailHandler *handlers.EmailHandler,
	oidcHandler *handlers.OIDCHandler,
) {
	// Apply global middleware
	s.router.Use(middleware.SecurityMiddleware())                         // Security headers
	s.router.Use(middleware.RateLimitMiddleware(s.rateLimiter))           // Rate limiting
	s.router.Use(middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB request size limit
	s.router.Use(middleware.MaliciousRouteInterceptor())                  // Block suspicious routes
	s.router.Use(middleware.CORSMiddleware())
	s.router.Use(middleware.RequestIDMiddleware())
	s.router.Use(middleware.AuditMiddleware(s.db, s.logger))

	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, models.APIResponse{
			Success: true,
			Message: "OK",
			Data: map[string]interface{}{
				"status":  "healthy",
				"version": version.Get(),
			},
		})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			// Authentication endpoints
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/register", accountHandler.Register)

			// Email verification (public)
			auth.GET("/email/verify", emailHandler.VerifyEmail)

			// Password reset endpoints (public)
			auth.GET("/password/request-reset", accountHandler.RequestPasswordReset)
			auth.POST("/password/reset", accountHandler.ResetPassword)

			// OIDC routes (public)
			oidc := auth.Group("/oidc")
			{
				oidc.GET("/providers", oidcHandler.ListPublicProviders)
				oidc.GET("/:provider/login", oidcHandler.InitiateLogin)
				oidc.GET("/:provider/callback", oidcHandler.HandleCallback)
			}
		}

		// Protected routes (all require authentication via JWT)
		// AuthMiddleware validates JWT token and sets user_id, username, and user_role in context
		// Handlers can safely use middleware.MustGetUserID() without error checking
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(s.jwtManager))
		protected.Use(middleware.RequireAuth())
		{
			// Additional auth routes (require authentication)
			protectedAuth := protected.Group("/auth")
			{
				protectedAuth.POST("/logout", authHandler.Logout)
			}
			// User routes
			users := protected.Group("/users")
			{
				// Self-service routes (all authenticated users)
				users.GET("/me/profile", userHandler.GetUserProfile)
				users.PUT("/me/profile", userHandler.UpdateUserProfile)
				users.GET("/me/email", userHandler.GetUserEmail)
				users.PUT("/me/email", userHandler.UpdateUserEmail)
				users.PUT("/me/security/password", userHandler.ChangePassword)

				// MFA routes
				users.POST("/me/security/mfa/enable", mfaHandler.EnableMFA)
				users.POST("/me/security/mfa/verify", mfaHandler.VerifyMFASetup)
				users.POST("/me/security/mfa/backup-codes", mfaHandler.GenerateBackupCodes)
				users.POST("/me/security/mfa/disable", mfaHandler.DisableMFA)

				// Power user and Admin routes
				users.GET("/check", userHandler.CheckUserExistence)

				// Admin routes
				users.GET("/search", userHandler.Search)
				users.GET("", userHandler.List)
				users.POST("", userHandler.Create)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
				users.GET("/:id/profile", userHandler.GetProfile)
				users.PUT("/:id/profile", userHandler.UpdateProfile)
				users.GET("/:id/email", userHandler.GetEmail)
				users.PUT("/:id/email", userHandler.UpdateEmail)
				users.PUT("/:id/security/password", userHandler.ChangePassword)
				users.POST("/:id/security/mfa/enable", mfaHandler.EnableMFA)
				users.POST("/:id/security/mfa/disable", mfaHandler.DisableMFA)
			}

			// Organization routes
			orgs := protected.Group("/organizations")
			{
				// All authenticated users can create organizations
				orgs.POST("", orgHandler.Create)

				// List organizations (supports ?all=true for admins)
				orgs.GET("", orgHandler.List)

				// Organization-specific routes
				orgs.DELETE("/:id", orgHandler.Delete)

				// Profile sub-resource
				orgs.GET("/:id/profile", orgHandler.GetProfile)
				orgs.PUT("/:id/profile", orgHandler.UpdateProfile)

				// Members sub-resource
				orgs.GET("/:id/members", orgHandler.ListMembers)
				orgs.POST("/:id/members", orgHandler.AddMembers)
				orgs.DELETE("/:id/members/:user_id", orgHandler.RemoveMember)
				orgs.PUT("/:id/members/:user_id", orgHandler.UpdateMemberRole)

				// Admins sub-resource
				orgs.GET("/:id/admins", orgHandler.ListAdmins)
				orgs.POST("/:id/admins/:user_id", orgHandler.PromoteToAdmin)
				orgs.DELETE("/:id/admins/:user_id", orgHandler.DemoteFromAdmin)

				// Owner sub-resource
				orgs.GET("/:id/owner", orgHandler.GetOwner)
				orgs.PUT("/:id/owner", orgHandler.TransferOwnership)

				// Quota sub-resource
				orgs.GET("/:id/quota", orgHandler.GetQuota)
				orgs.PUT("/:id/quota", orgHandler.UpdateQuota)
			}

			// Template routes
			templates := protected.Group("/templates")
			{
				// List templates (supports ?all=true for admins)
				templates.GET("", templateHandler.List)

				// Template sub-resource routes
				templates.GET("/:id/profile", templateHandler.GetProfile)
				templates.PUT("/:id/profile", templateHandler.UpdateProfile)
				templates.GET("/:id/image-ref", templateHandler.GetImageRef)
				templates.PUT("/:id/image-ref", templateHandler.UpdateImageRef)
				templates.GET("/:id/template-yaml", templateHandler.GetTemplateYAML)
				templates.PUT("/:id/template-yaml", templateHandler.UpdateTemplateYAML)
				templates.GET("/:id/public", templateHandler.GetPublic)
				templates.PUT("/:id/public", templateHandler.SetPublic)
				templates.GET("/:id/quota", templateHandler.GetQuota)
				templates.PUT("/:id/quota", templateHandler.UpdateQuota)
			}

			// Workspace routes
			workspaces := protected.Group("/workspaces")
			{
				// List and Create workspaces
				workspaces.GET("", workspaceHandler.List)
				workspaces.POST("", workspaceHandler.Create)
				workspaces.DELETE("/:id", workspaceHandler.Delete)

				// Workspace sub-resource routes
				workspaces.GET("/:id/profile", workspaceHandler.GetProfile)
				workspaces.PUT("/:id/profile", workspaceHandler.UpdateProfile)
				workspaces.GET("/:id/template", workspaceHandler.GetTemplate)
				workspaces.GET("/:id/volumes", workspaceHandler.ListVolumeMounts)
				workspaces.PUT("/:id/volumes", workspaceHandler.UpdateVolumeMounts)
				workspaces.GET("/:id/quota", workspaceHandler.GetQuota)
				workspaces.PUT("/:id/quota", workspaceHandler.UpdateQuota)
				workspaces.GET("/:id/public", workspaceHandler.GetPublic)
				workspaces.PUT("/:id/public", workspaceHandler.SetPublic)
				workspaces.GET("/:id/owner", workspaceHandler.GetOwner)
				workspaces.PUT("/:id/owner", workspaceHandler.TransferOwnership)
				workspaces.GET("/:id/state", workspaceHandler.GetCurrentState)
				workspaces.PUT("/:id/state", workspaceHandler.UpdateTargetState)
			}

			// Workspace transfer routes (kept for backward compatibility)
			workspaceTransfers := protected.Group("/workspace-transfers")
			{
				workspaceTransfers.GET("/pending", workspaceHandler.ListPendingTransfers)
				workspaceTransfers.GET("/:transfer_id", workspaceHandler.GetTransfer)
				workspaceTransfers.POST("/:transfer_id/respond", workspaceHandler.RespondToTransfer)
				workspaceTransfers.POST("/:transfer_id/cancel", workspaceHandler.CancelTransfer)
			}

			// Volume routes
			volumes := protected.Group("/volumes")
			{
				// List and Create volumes
				volumes.GET("", volumeHandler.List)
				volumes.POST("", volumeHandler.Create)
				volumes.DELETE("/:id", volumeHandler.Delete)

				// Volume sub-resource routes
				volumes.GET("/:id/profile", volumeHandler.GetProfile)
				volumes.PUT("/:id/profile", volumeHandler.UpdateProfile)
				volumes.GET("/:id/size-mb", volumeHandler.GetSizeMB)
				volumes.PUT("/:id/size-mb", volumeHandler.UpdateSizeMB)
				volumes.GET("/:id/storage-class", volumeHandler.GetStorageClass)
				volumes.GET("/:id/access-mode", volumeHandler.GetAccessMode)
				volumes.GET("/:id/owner", volumeHandler.GetOwner)
				volumes.PUT("/:id/owner", volumeHandler.TransferOwnership)
				volumes.GET("/:id/public", volumeHandler.GetPublic)
				volumes.PUT("/:id/public", volumeHandler.SetPublic)
			}

			// Settings routes (admin only - checked in handler)
			settings := protected.Group("/settings")
			{
				// General settings
				settings.GET("", settingHandler.GetAllSettings)
				settings.PUT("", settingHandler.BatchUpdateSettings)

				// Key-value settings
				settings.GET("/kv/:key", settingHandler.GetSetting)
				settings.PUT("/kv/:key", settingHandler.UpdateSetting)

				// OIDC provider settings
				settings.POST("/oidc", oidcHandler.CreateProvider)
				settings.GET("/oidc", oidcHandler.ListProviders)
				settings.PUT("/oidc/:id", oidcHandler.UpdateProvider)
				settings.DELETE("/oidc/:id", oidcHandler.DeleteProvider)
			}
		}
	}
}

// GetRouter returns the gin router
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Run starts the HTTP server with proper timeouts and limits
func (s *Server) Run(address string) error {
	// Create HTTP server with security configurations
	s.httpServer = &http.Server{
		Addr:    address,
		Handler: s.router,

		// Timeout configurations
		ReadTimeout:       15 * time.Second,  // Time to read request headers and body
		ReadHeaderTimeout: 10 * time.Second,  // Time to read request headers only
		WriteTimeout:      30 * time.Second,  // Time to write response
		IdleTimeout:       120 * time.Second, // Keep-alive timeout

		// Size limits
		MaxHeaderBytes: 1 << 20, // 1MB max header size
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
