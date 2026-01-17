package api

import (
	"context"
	"net/http"
	"time"

	_ "github.com/davidliyutong/idekube-controller/docs/api"
	"github.com/davidliyutong/idekube-controller/internal/handlers"
	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/permission"
	"github.com/davidliyutong/idekube-controller/internal/version"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Server represents the API server
type Server struct {
	router            *gin.Engine
	httpServer        *http.Server
	db                *gorm.DB
	jwtManager        *middleware.JWTManager
	permissionService *permission.PermissionService
	rateLimiter       *middleware.RateLimiter
	logger            *logger.Logger
}

// NewServer creates a new API server
func NewServer(db *gorm.DB, jwtManager *middleware.JWTManager, permissionService *permission.PermissionService, logger *logger.Logger) *Server {
	// Create rate limiter: 100 requests per minute per IP
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         20,
	})

	return &Server{
		router:            gin.Default(),
		db:                db,
		jwtManager:        jwtManager,
		permissionService: permissionService,
		rateLimiter:       rateLimiter,
		logger:            logger,
	}
}

// SetupRoutes sets up all API routes
func (s *Server) SetupRoutes(
	userHandler *handlers.UserHandler,
	mfaHandler *handlers.MFAHandler,
	orgHandler *handlers.OrganizationHandler,
	templateHandler *handlers.TemplateHandler,
	workspaceHandler *handlers.WorkspaceHandler,
	volumeHandler *handlers.VolumeHandler,
	permissionHandler *handlers.PermissionHandler,
	settingHandler *handlers.SettingHandler,
	apiKeyHandler *handlers.APIKeyHandler,
	emailHandler *handlers.EmailHandler,
	oidcHandler *handlers.OIDCHandler,
	webhookHandler *handlers.WebhookHandler,
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
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/register", userHandler.Register)
			
			// Email verification and password reset (public)
			auth.GET("/verify-email", emailHandler.VerifyEmail)
			auth.POST("/request-password-reset", emailHandler.RequestPasswordReset)
			auth.POST("/reset-password", emailHandler.ResetPassword)
			
			// OIDC routes (public)
			auth.GET("/oidc/:provider/login", oidcHandler.InitiateLogin)
			auth.GET("/oidc/:provider/callback", oidcHandler.HandleCallback)
		}

		// Protected routes (all require authentication via JWT)
		// AuthMiddleware validates JWT token and sets user_id, username, and user_role in context
		// Handlers can safely use middleware.MustGetUserID() without error checking
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(s.jwtManager))
		protected.Use(middleware.RequireAuth())
		protected.Use(middleware.RBACMiddleware(s.permissionService))
		{
			// Additional auth routes (require authentication)
			protectedAuth := protected.Group("/auth")
			{
				protectedAuth.POST("/logout", userHandler.Logout)
			}
			// User routes
			users := protected.Group("/users")
			users.Use(middleware.RBACCheckEndpoint(s.permissionService, "user"))
			{
				// Self-service routes (all authenticated users)
				users.GET("/me", userHandler.GetProfile)
				users.PUT("/me", userHandler.UpdateProfile)
				users.POST("/me/password", userHandler.ChangePassword)
				users.GET("/:id", userHandler.GetUser)

				// MFA routes
				// FIXME: Uncomment once MFAHandler is properly integrated
				// users.POST("/me/mfa/enable", mfaHandler.EnableMFA)
				// users.POST("/me/mfa/verify", mfaHandler.VerifyMFASetup)
				// users.POST("/me/mfa/backup-codes", mfaHandler.GenerateBackupCodes)
				// users.POST("/me/mfa/disable", mfaHandler.DisableMFA)

				// Power user and Admin routes
				users.GET("/check", userHandler.CheckUserExists)

				// Admin routes
				users.GET("/search", userHandler.SearchUsers)
				users.GET("", userHandler.ListUsers)
				users.POST("", userHandler.CreateUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)

				// Role management (admin only)
				users.POST("/:id/roles", permissionHandler.AssignRole)
				users.DELETE("/:id/roles", permissionHandler.RemoveRole)
				users.GET("/:id/roles", permissionHandler.GetUserRoles)
			}

			// Organization routes
			orgs := protected.Group("/organizations")
			orgs.Use(middleware.RBACCheckEndpoint(s.permissionService, "organization"))
			{
				// All authenticated users can create organizations
				orgs.POST("", orgHandler.CreateOrganization)

				// List organizations (supports ?all=true for admins)
				orgs.GET("", orgHandler.ListUserOrganizations)

				// Organization-specific routes
				orgs.GET("/:id", orgHandler.GetOrganization)
				orgs.PUT("/:id", orgHandler.UpdateOrganization)
				orgs.DELETE("/:id", orgHandler.DeleteOrganization)

				// Member management (requires org admin)
				orgs.POST("/:id/members", orgHandler.AddMember)
				orgs.DELETE("/:id/members/:user_id", orgHandler.RemoveMember)
				orgs.PUT("/:id/members/:user_id", orgHandler.UpdateMemberRole)

				// Admin role management (requires org owner)
				orgs.POST("/:id/admins/:user_id", orgHandler.PromoteToAdmin)
				orgs.DELETE("/:id/admins/:user_id", orgHandler.DemoteFromAdmin)

				// User search for invitations (requires org admin)
				orgs.GET("/:id/search-users", orgHandler.SearchUsers)
			}

			// Template routes
			templates := protected.Group("/templates")
			templates.Use(middleware.RBACCheckEndpoint(s.permissionService, "template"))
			{
				// List templates (supports ?all=true for admins)
				templates.GET("", templateHandler.ListTemplates)

				// Template CRUD
				templates.POST("", templateHandler.CreateTemplate)
				templates.GET("/:id", templateHandler.GetTemplate)
				templates.PUT("/:id", templateHandler.UpdateTemplate)
				templates.DELETE("/:id", templateHandler.DeleteTemplate)
			}

			// Workspace routes
			workspaces := protected.Group("/workspaces")
			workspaces.Use(middleware.RBACCheckEndpoint(s.permissionService, "workspace"))
			{
				// List workspaces (supports ?organization_id= filter)
				workspaces.GET("", workspaceHandler.ListWorkspaces)

				// Workspace CRUD
				workspaces.POST("", workspaceHandler.CreateWorkspace)
				workspaces.GET("/:id", workspaceHandler.GetWorkspace)
				workspaces.PUT("/:id", workspaceHandler.UpdateWorkspace)
				workspaces.DELETE("/:id", workspaceHandler.DeleteWorkspace)

				// Workspace operations
				workspaces.POST("/:id/start", workspaceHandler.StartWorkspace)
				workspaces.POST("/:id/stop", workspaceHandler.StopWorkspace)

				// Volume management
				workspaces.POST("/:id/volumes/:volume_id", workspaceHandler.AttachVolume)
				workspaces.DELETE("/:id/volumes/:volume_id", workspaceHandler.DetachVolume)

				// Workspace transfer
				workspaces.POST("/:id/transfer", workspaceHandler.InitiateTransfer)
			}

			// Workspace transfer routes
			workspaceTransfers := protected.Group("/workspace-transfers")
			workspaceTransfers.Use(middleware.RBACCheckEndpoint(s.permissionService, "workspace"))
			{
				workspaceTransfers.GET("/pending", workspaceHandler.ListPendingTransfers)
				workspaceTransfers.GET("/:transfer_id", workspaceHandler.GetTransfer)
				workspaceTransfers.POST("/:transfer_id/respond", workspaceHandler.RespondToTransfer)
				workspaceTransfers.POST("/:transfer_id/cancel", workspaceHandler.CancelTransfer)
			}

			// Volume routes
			volumes := protected.Group("/volumes")
			volumes.Use(middleware.RBACCheckEndpoint(s.permissionService, "volume"))
			{
				// List volumes (supports ?organization_id= filter)
				volumes.GET("", volumeHandler.ListVolumes)

				// Volume CRUD
				volumes.POST("", volumeHandler.CreateVolume)
				volumes.GET("/:id", volumeHandler.GetVolume)
				volumes.PUT("/:id", volumeHandler.UpdateVolume)
				volumes.DELETE("/:id", volumeHandler.DeleteVolume)

				// Volume sync
				volumes.POST("/:id/sync", volumeHandler.SyncVolumeStatus)
			}

			// Permission and policy management routes (admin only)
			permissions := protected.Group("/permissions")
			{
				permissions.POST("/check", permissionHandler.CheckPermission)
			}

			policies := protected.Group("/policies")
			{
				policies.GET("", permissionHandler.GetAllPolicies)
				policies.POST("", permissionHandler.AddPolicy)
				policies.DELETE("", permissionHandler.RemovePolicy)
			}

			// Settings routes (admin only via RBAC)
			settings := protected.Group("/settings")
			settings.Use(middleware.RBACCheckEndpoint(s.permissionService, "settings"))
			{
				settings.GET("", settingHandler.GetAllSettings)
				settings.PUT("", settingHandler.BatchUpdateSettings)
				settings.GET("/:key", settingHandler.GetSetting)
				settings.PUT("/:key", settingHandler.UpdateSetting)
			}

			// API Key routes (authenticated users)
			apiKeys := protected.Group("/api-keys")
			apiKeys.Use(middleware.RBACCheckEndpoint(s.permissionService, "api_key"))
			{
				apiKeys.POST("", apiKeyHandler.CreateAPIKey)
				apiKeys.GET("", apiKeyHandler.ListAPIKeys)
				apiKeys.GET("/:id", apiKeyHandler.GetAPIKey)
				apiKeys.DELETE("/:id", apiKeyHandler.RevokeAPIKey)
			}

			// Webhook routes (authenticated users)
			webhooks := protected.Group("/webhooks")
			webhooks.Use(middleware.RBACCheckEndpoint(s.permissionService, "webhook"))
			{
				webhooks.POST("", webhookHandler.CreateWebhook)
				webhooks.GET("", webhookHandler.ListWebhooks)
				webhooks.GET("/:id", webhookHandler.GetWebhook)
				webhooks.PUT("/:id", webhookHandler.UpdateWebhook)
				webhooks.DELETE("/:id", webhookHandler.DeleteWebhook)
				webhooks.POST("/:id/test", webhookHandler.TestWebhook)
			}

			// OIDC Provider management routes (admin only)
			oidcProviders := protected.Group("/oidc/providers")
			oidcProviders.Use(middleware.RBACCheckEndpoint(s.permissionService, "oidc_provider"))
			{
				oidcProviders.POST("", oidcHandler.CreateProvider)
				oidcProviders.GET("", oidcHandler.ListProviders)
				oidcProviders.PUT("/:id", oidcHandler.UpdateProvider)
				oidcProviders.DELETE("/:id", oidcHandler.DeleteProvider)
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
