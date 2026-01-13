package api

import (
	"github.com/davidliyutong/idekube-controller/internal/handlers"
	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Server represents the API server
type Server struct {
	router     *gin.Engine
	db         *gorm.DB
	jwtManager *middleware.JWTManager
}

// NewServer creates a new API server
func NewServer(db *gorm.DB, jwtManager *middleware.JWTManager) *Server {
	return &Server{
		router:     gin.Default(),
		db:         db,
		jwtManager: jwtManager,
	}
}

// SetupRoutes sets up all API routes
func (s *Server) SetupRoutes(
	userHandler *handlers.UserHandler,
	orgHandler *handlers.OrganizationHandler,
	templateHandler *handlers.TemplateHandler,
	workspaceHandler *handlers.WorkspaceHandler,
	volumeHandler *handlers.VolumeHandler,
) {
	// Apply global middleware
	s.router.Use(middleware.CORSMiddleware())
	s.router.Use(middleware.RequestIDMiddleware())
	s.router.Use(middleware.AuditMiddleware(s.db))
	
	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, models.APIResponse{
			Success: true,
			Message: "OK",
			Data: map[string]interface{}{
				"status":  "healthy",
				"version": "v1.0.0",
			},
		})
	})
	
	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/register", userHandler.Register)
			// TODO: Add OIDC routes
			// auth.GET("/oidc/login", oidcHandler.Login)
			// auth.GET("/oidc/callback", oidcHandler.Callback)
		}
		
		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(s.jwtManager))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetProfile)
				users.POST("/me/password", userHandler.ChangePassword)
				users.GET("/:id", userHandler.GetUser)
				
				// Admin only routes
				adminUsers := users.Group("")
				adminUsers.Use(middleware.RequireRole(models.UserRoleAdmin, models.UserRoleSuperAdmin))
				{
					adminUsers.GET("", userHandler.ListUsers)
					adminUsers.PUT("/:id", userHandler.UpdateUser)
					adminUsers.DELETE("/:id", userHandler.DeleteUser)
				}
			}
			
			// Organization routes
			orgs := protected.Group("/organizations")
			{
				orgs.POST("", orgHandler.CreateOrganization)
				orgs.GET("", orgHandler.ListUserOrganizations)
				orgs.GET("/:id", orgHandler.GetOrganization)
				orgs.PUT("/:id", orgHandler.UpdateOrganization)
				orgs.DELETE("/:id", orgHandler.DeleteOrganization)
				orgs.POST("/:id/members", orgHandler.AddMember)
				orgs.DELETE("/:id/members/:user_id", orgHandler.RemoveMember)
				orgs.PUT("/:id/members/:user_id", orgHandler.UpdateMemberRole)
			}
			
			// Template routes
			templates := protected.Group("/templates")
			{
				templates.GET("", templateHandler.ListTemplates)
				templates.POST("", templateHandler.CreateTemplate)
				templates.GET("/:id", templateHandler.GetTemplate)
				templates.PUT("/:id", templateHandler.UpdateTemplate)
				templates.DELETE("/:id", templateHandler.DeleteTemplate)
			}
			
			// Workspace routes
			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", workspaceHandler.ListWorkspaces)
				workspaces.POST("", workspaceHandler.CreateWorkspace)
				workspaces.GET("/:id", workspaceHandler.GetWorkspace)
				workspaces.PUT("/:id", workspaceHandler.UpdateWorkspace)
				workspaces.DELETE("/:id", workspaceHandler.DeleteWorkspace)
				workspaces.POST("/:id/start", workspaceHandler.StartWorkspace)
				workspaces.POST("/:id/stop", workspaceHandler.StopWorkspace)
				workspaces.POST("/:id/volumes/:volume_id", workspaceHandler.AttachVolume)
				workspaces.DELETE("/:id/volumes/:volume_id", workspaceHandler.DetachVolume)
			}
			
			// Volume routes
			volumes := protected.Group("/volumes")
			{
				volumes.GET("", volumeHandler.ListVolumes)
				volumes.POST("", volumeHandler.CreateVolume)
				volumes.GET("/:id", volumeHandler.GetVolume)
				volumes.PUT("/:id", volumeHandler.UpdateVolume)
				volumes.DELETE("/:id", volumeHandler.DeleteVolume)
				volumes.POST("/:id/sync", volumeHandler.SyncVolumeStatus)
			}
		}
	}
}

// GetRouter returns the gin router
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Run starts the HTTP server
func (s *Server) Run(address string) error {
	return s.router.Run(address)
}
