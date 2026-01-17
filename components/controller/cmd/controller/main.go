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
	"github.com/davidliyutong/idekube-controller/internal/opa"
	"github.com/davidliyutong/idekube-controller/internal/permission"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/davidliyutong/idekube-controller/pkg/database"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	redisClient "github.com/davidliyutong/idekube-controller/pkg/redis"
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
	serverAddr     string
	skipMigrations bool
	migrationsPath string
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
	// Load configuration first to get LOG_LEVEL
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize zap logger with LOG_LEVEL from config
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Parse and set log level from config
	logLevel, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = zapcore.InfoLevel // Default to info level if invalid
	}
	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting idekube-controller", zap.String("log_level", cfg.LogLevel))

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

	// Configure GORM logger with zap integration and LOG_LEVEL
	gormDB.Logger = config.NewGormLogger(zapLogger)

	// Initialize Redis connection
	redisConfig := &redisClient.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	redis, err := redisClient.NewClient(redisConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redis.Close()
	zapLogger.Info("Redis client initialized", zap.String("host", cfg.Redis.Host), zap.Int("port", cfg.Redis.Port))

	// Initialize Redis queue client
	redisQueue, err := queue.NewRedisQueueClient(redis.GetClient(), zapLogger)
	if err != nil {
		return fmt.Errorf("failed to create Redis queue client: %w", err)
	}
	defer redisQueue.Close()
	zapLogger.Info("Redis queue client initialized")

	// Initialize OPA enforcer and permission service
	log := logger.NewLogger(zapLogger)
	opaEnforcer, err := opa.NewEnforcer(gormDB, cfg.OPA.PolicyPath, cfg.OPA.DataPath, log)
	if err != nil {
		return fmt.Errorf("failed to initialize OPA enforcer: %w", err)
	}
	permissionService := permission.NewPermissionService(opaEnforcer, log)
	resourcePermissionService := permission.NewResourcePermissionService(permissionService)
	zapLogger.Info("OPA enforcer and permission service initialized",
		zap.String("policy_path", cfg.OPA.PolicyPath),
		zap.String("data_path", cfg.OPA.DataPath))

	// Initialize repositories
	userRepo := repository.NewUserRepository(gormDB)
	orgRepo := repository.NewOrganizationRepository(gormDB)
	templateRepo := repository.NewTemplateRepository(gormDB)
	workspaceRepo := repository.NewWorkspaceRepository(gormDB)
	volumeRepo := repository.NewVolumeRepository(gormDB)
	quotaRepo := repository.NewQuotaRepository(gormDB)
	workspaceTransferRepo := repository.NewWorkspaceTransferRepository(gormDB)
	settingRepo := repository.NewSettingRepository(gormDB)
	apiKeyRepo := repository.NewAPIKeyRepository(gormDB)
	webhookRepo := repository.NewWebhookRepository(gormDB)
	oidcProviderRepo := repository.NewOIDCProviderRepository(gormDB)
	oauthSessionRepo := repository.NewOAuthSessionRepository(gormDB)

	// Initialize message queue event publisher
	eventPublisher, err := queue.NewEventPublisher(redisQueue, zapLogger)
	if err != nil {
		return fmt.Errorf("failed to create event publisher: %w", err)
	}

	// Initialize setting service first (needed for JWT configuration)
	settingService := services.NewSettingService(settingRepo, redis)

	// Setup context for fetching settings
	initCtx := context.Background()

	// Fetch JWT configuration from database settings
	accessTokenMinutes := settingService.GetSettingAsInt(initCtx, "access_token_expiration_minutes", 15)
	refreshTokenDays := settingService.GetSettingAsInt(initCtx, "refresh_token_expiration_days", 30)
	zapLogger.Info("JWT configuration loaded from database",
		zap.Int("access_token_minutes", accessTokenMinutes),
		zap.Int("refresh_token_days", refreshTokenDays))

	// Initialize JWT manager with configuration from database
	jwtManager := middleware.NewJWTManager(&middleware.JWTConfig{
		SecretKey:            cfg.JWTSecret,
		AccessTokenDuration:  time.Duration(accessTokenMinutes) * time.Minute,
		RefreshTokenDuration: time.Duration(refreshTokenDays) * 24 * time.Hour,
		RedisClient:          redis,
	})

	// Initialize login attempt service (needs settingService)
	loginAttemptService := services.NewLoginAttemptService(redis, settingService)

	userService := services.NewUserService(userRepo, jwtManager, eventPublisher, loginAttemptService)
	mfaService := services.NewMFAService(userRepo)
	templateService := services.NewTemplateService(templateRepo, orgRepo)
	volumeService := services.NewVolumeService(volumeRepo, eventPublisher, zapLogger, resourcePermissionService)
	workspaceService := services.NewWorkspaceService(workspaceRepo, templateRepo, volumeRepo, eventPublisher, zapLogger, resourcePermissionService)
	workspaceTransferService := services.NewWorkspaceTransferService(workspaceTransferRepo, workspaceRepo, userRepo)
	quotaService := services.NewQuotaService(quotaRepo, workspaceRepo, volumeRepo)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)
	webhookService := services.NewWebhookService(webhookRepo)
	oidcService := services.NewOIDCService(oidcProviderRepo, userRepo, oauthSessionRepo)
	
	// EmailService with SMTP configuration
	emailService := services.NewEmailService(
		userRepo,
		oauthSessionRepo,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password,
		cfg.SMTP.UseTLS,
		cfg.SMTP.FromEmail,
		cfg.SMTP.FromName,
		cfg.BaseURL,
	)

	// OrganizationService needs UserService for SearchUsersForInvite and ResourcePermissionService
	orgService := services.NewOrganizationService(orgRepo, userRepo, userService, eventPublisher, resourcePermissionService)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize status consumer to receive updates from housekeeper
	statusConsumer, err := queue.NewStatusConsumer(redisQueue, workspaceRepo, zapLogger)
	if err != nil {
		return fmt.Errorf("failed to create status consumer: %w", err)
	}

	// Start status consumer in a goroutine
	go func() {
		if err := statusConsumer.Start(ctx); err != nil {
			zapLogger.Error("Status consumer error", zap.Error(err))
		}
	}()

	// Initialize admin account if ADMIN_PASSWORD is set
	if cfg.AdminPassword != "" {
		if err := initializeAdminAccount(ctx, userService, permissionService, cfg.AdminPassword, zapLogger); err != nil {
			zapLogger.Error("Failed to initialize admin account", zap.Error(err))
			// Don't fail startup, just log the error
		}
	}

	// Initialize default settings
	zapLogger.Info("Initializing default settings...")
	if err := settingService.InitializeDefaultSettings(ctx); err != nil {
		zapLogger.Error("Failed to initialize default settings", zap.Error(err))
		// Don't fail startup, just log the error
	}

	// Note: quotaService will be integrated in Phase 2 for resource quota enforcement
	_ = quotaService

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, cfg.EnableRegistration)
	mfaHandler := handlers.NewMFAHandler(mfaService)
	orgHandler := handlers.NewOrganizationHandler(orgService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	volumeHandler := handlers.NewVolumeHandler(volumeService)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService, workspaceTransferService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	settingHandler := handlers.NewSettingHandler(settingService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	emailHandler := handlers.NewEmailHandler(emailService)
	oidcHandler := handlers.NewOIDCHandler(oidcService, userService, jwtManager)
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Create and setup API server
	server := api.NewServer(gormDB, jwtManager, permissionService, log)
	server.SetupRoutes(
		userHandler,
		mfaHandler,
		orgHandler,
		templateHandler,
		workspaceHandler,
		volumeHandler,
		permissionHandler,
		settingHandler,
		apiKeyHandler,
		emailHandler,
		oidcHandler,
		webhookHandler,
	)

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

// initializeAdminAccount creates the admin account if it doesn't exist and assigns super_admin role
func initializeAdminAccount(ctx context.Context, userService *services.UserService, permissionService *permission.PermissionService, password string, logger *zap.Logger) error {
	// Check if admin user already exists
	user, err := userService.GetUserByUsername(ctx, "admin")
	if err == nil {
		logger.Info("Admin account already exists, checking role binding...")

		// Check if admin already has super_admin role binding
		roles, err := permissionService.GetUserRoles(ctx, user.ID)
		if err != nil {
			logger.Warn("Failed to get admin roles", zap.Error(err))
		} else {
			hasSuperAdmin := false
			for _, role := range roles {
				if role == "role:super_admin" {
					hasSuperAdmin = true
					break
				}
			}

			if !hasSuperAdmin {
				logger.Info("Assigning super_admin role to admin user...")
				if err := permissionService.AssignRole(ctx, user.ID, "role:super_admin"); err != nil {
					logger.Error("Failed to assign super_admin role to admin", zap.Error(err))
				} else {
					logger.Info("Super_admin role assigned to admin successfully")
				}
			} else {
				logger.Info("Admin user already has super_admin role")
			}
		}

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

	createdUser, err := userService.CreateUser(ctx, adminUser)
	if err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}

	logger.Info("Admin account created successfully",
		zap.String("username", "admin"),
		zap.String("email", email),
		zap.Int64("user_id", createdUser.ID))

	// Assign super_admin role via OPA
	if err := permissionService.AssignRole(ctx, createdUser.ID, "role:super_admin"); err != nil {
		logger.Error("Failed to assign super_admin role to newly created admin", zap.Error(err))
	} else {
		logger.Info("Super_admin role assigned to admin successfully")
	}

	logger.Warn("IMPORTANT: Please change the admin password after first login!")

	return nil
}