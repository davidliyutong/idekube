package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/davidliyutong/idekube-controller/docs/api" // Import generated swagger docs
	"github.com/davidliyutong/idekube-controller/internal/api"
	"github.com/davidliyutong/idekube-controller/internal/config"
	"github.com/davidliyutong/idekube-controller/internal/handlers"
	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/models"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/davidliyutong/idekube-controller/pkg/database"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	"github.com/davidliyutong/idekube-controller/pkg/rbac"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// @title IDEKube Controller API
// @version 1.0
// @description API服务器，用于管理云IDE平台的工作区、模板、用户和组织
// @termsOfService https://github.com/davidliyutong/idekube

// @contact.name API Support
// @contact.url https://github.com/davidliyutong/idekube/issues
// @contact.email support@idekube.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var (
	serverAddr      string
	skipMigrations  bool
	migrationsPath  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "idekube-controller",
		Short: "IDEKube Controller - API server for cloud IDE platform",
		Long:  "IDEKube Controller manages workspaces, templates, users, and organizations for the cloud IDE platform",
		RunE:  runController,
	}

	// Add flags
	rootCmd.Flags().StringVarP(&serverAddr, "addr", "a", ":8080", "Server address to listen on")
	rootCmd.Flags().BoolVar(&skipMigrations, "skip-migrations", false, "Skip database migrations on startup")
	rootCmd.Flags().StringVar(&migrationsPath, "migrations-path", "./migrations", "Path to database migrations")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runController(cmd *cobra.Command, args []string) error {
	// Initialize zap logger
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting idekube-controller")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override server address if provided via flag
	if serverAddr != "" {
		cfg.ServerAddress = serverAddr
	}

	// Initialize PostgreSQL connection
	db, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// Run database migrations unless skipped
	if !skipMigrations {
		zapLogger.Info("Running database migrations...")
		migrationConfig := database.MigrationConfig{
			MigrationsPath: migrationsPath,
			DatabaseName:   cfg.Postgres.Database,
		}
		if err := database.RunMigrations(db, migrationConfig); err != nil {
			return fmt.Errorf("failed to run database migrations: %w", err)
		}
		zapLogger.Info("Database migrations completed successfully")
	} else {
		zapLogger.Info("Skipping database migrations")
	}

	// Get GORM DB instance
	gormDB := db.DB()

	// Initialize RabbitMQ connection
	mqClient, err := queue.NewRabbitMQClient(cfg.RabbitMQ)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer mqClient.Close()

	// Initialize JWT manager
	jwtManager := middleware.NewJWTManager(&middleware.JWTConfig{
		SecretKey:     cfg.JWTSecret,
		TokenDuration: time.Duration(cfg.JWTExpirationHours) * time.Hour,
	})

	// Initialize RBAC client
	if cfg.RBACEndpoint == "" {
		return fmt.Errorf("RBAC_ENDPOINT is required but not configured")
	}
	rbacClient := rbac.NewClient(cfg.RBACEndpoint)
	zapLogger.Info("RBAC client initialized", zap.String("endpoint", cfg.RBACEndpoint))

	// Initialize repositories
	userRepo := repository.NewUserRepository(gormDB)
	orgRepo := repository.NewOrganizationRepository(gormDB)
	templateRepo := repository.NewTemplateRepository(gormDB)
	workspaceRepo := repository.NewWorkspaceRepository(gormDB)
	volumeRepo := repository.NewVolumeRepository(gormDB)
	quotaRepo := repository.NewQuotaRepository(gormDB)

	// Initialize message queue event publisher
	eventPublisher, err := queue.NewEventPublisher(mqClient, zapLogger)
	if err != nil {
		return fmt.Errorf("failed to create event publisher: %w", err)
	}

	// Initialize services
	userService := services.NewUserService(userRepo, jwtManager, eventPublisher)
	templateService := services.NewTemplateService(templateRepo, orgRepo)
	volumeService := services.NewVolumeService(volumeRepo, eventPublisher, zapLogger)
	workspaceService := services.NewWorkspaceService(workspaceRepo, templateRepo, volumeRepo, eventPublisher, zapLogger)
	quotaService := services.NewQuotaService(quotaRepo, workspaceRepo, volumeRepo)

	// OrganizationService needs UserService for SearchUsersForInvite
	orgService := services.NewOrganizationService(orgRepo, userRepo, userService, eventPublisher)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize admin account if ADMIN_PASSWORD is set
	if cfg.AdminPassword != "" {
		if err := initializeAdminAccount(ctx, userService, cfg.AdminPassword, zapLogger); err != nil {
			zapLogger.Error("Failed to initialize admin account", zap.Error(err))
			// Don't fail startup, just log the error
		}
	}

	// Note: quotaService will be integrated in Phase 2 for resource quota enforcement
	_ = quotaService

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	orgHandler := handlers.NewOrganizationHandler(orgService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	volumeHandler := handlers.NewVolumeHandler(volumeService)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)

	// Create and setup API server
	server := api.NewServer(gormDB, jwtManager, rbacClient)
	server.SetupRoutes(userHandler, orgHandler, templateHandler, workspaceHandler, volumeHandler)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start API server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		addr := cfg.ServerAddress
		if addr == "" {
			addr = ":8080"
		}
		zapLogger.Info("Starting API server", zap.String("addr", addr))
		if err := server.Run(addr); err != nil {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	zapLogger.Info("Controller started successfully")

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		zapLogger.Info("Shutting down gracefully...")
	case err := <-errChan:
		zapLogger.Error("Server error", zap.Error(err))
		return err
	}

	cancel()
	zapLogger.Info("Controller stopped")
	return nil
}

// initializeAdminAccount creates the admin account if it doesn't exist
func initializeAdminAccount(ctx context.Context, userService *services.UserService, password string, logger *zap.Logger) error {
	// Check if admin user already exists
	_, err := userService.GetUserByUsername(ctx, "admin")
	if err == nil {
		logger.Info("Admin account already exists, skipping initialization")
		return nil
	}

	// Admin user doesn't exist, create it
	logger.Info("Creating admin account...")
	
	email := "admin@idekube.local"
	displayName := "System Administrator"
	
	adminUser := &models.CreateUserRequest{
		Username:    "admin",
		Email:       &email,
		Password:    password,
		Role:        models.UserRoleSuperAdmin,
		DisplayName: &displayName,
	}

	_, err = userService.CreateUser(ctx, adminUser)
	if err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}

	logger.Info("Admin account created successfully",
		zap.String("username", "admin"),
		zap.String("email", email))
	logger.Warn("IMPORTANT: Please change the admin password after first login!")

	return nil
}

