package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davidliyutong/idekube-controller/internal/api"
	"github.com/davidliyutong/idekube-controller/internal/config"
	"github.com/davidliyutong/idekube-controller/internal/handlers"
	"github.com/davidliyutong/idekube-controller/internal/middleware"
	"github.com/davidliyutong/idekube-controller/internal/repository"
	"github.com/davidliyutong/idekube-controller/internal/services"
	"github.com/davidliyutong/idekube-controller/pkg/database"
	"github.com/davidliyutong/idekube-controller/pkg/k8s"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
	"github.com/davidliyutong/idekube-controller/pkg/queue"
	
	_ "github.com/davidliyutong/idekube-controller/docs" // Import generated swagger docs
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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting idekube-controller")
	
	// Initialize zap logger for services
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLogger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}
	defer zapLogger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kubernetes client
	k8sClientset, err := k8s.NewClientset(cfg.Kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	
	k8sNamespace := cfg.Namespace
	if k8sNamespace == "" {
		k8sNamespace = "idekube"
	}

	// Initialize PostgreSQL connection
	db, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Run database migrations
	log.Info("Running database migrations...")
	migrationConfig := database.MigrationConfig{
		MigrationsPath: "./migrations",
		DatabaseName:   cfg.Postgres.Database,
	}
	if err := database.RunMigrations(db, migrationConfig); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Info("Database migrations completed successfully")

	// Get GORM DB instance
	gormDB := db.DB()

	// Initialize RabbitMQ connection
	mqClient, err := queue.NewRabbitMQClient(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer mqClient.Close()

	// Initialize JWT manager
	jwtManager := middleware.NewJWTManager(&middleware.JWTConfig{
		SecretKey:     cfg.JWTSecret,
		TokenDuration: time.Duration(cfg.JWTExpirationHours) * time.Hour,
	})

	// Initialize repositories
	userRepo := repository.NewUserRepository(gormDB)
	orgRepo := repository.NewOrganizationRepository(gormDB)
	templateRepo := repository.NewTemplateRepository(gormDB)
	workspaceRepo := repository.NewWorkspaceRepository(gormDB)
	volumeRepo := repository.NewVolumeRepository(gormDB)
	quotaRepo := repository.NewQuotaRepository(gormDB)

	// Initialize K8s managers
	pvcManager := k8s.NewPVCManager(k8sClientset, k8sNamespace)
	deploymentManager := k8s.NewDeploymentManager(k8sClientset, k8sNamespace)
	serviceManager := k8s.NewServiceManager(k8sClientset, k8sNamespace)
	
	// Note: K8S managers will be deprecated after HouseKeeper integration
	_ = pvcManager
	_ = deploymentManager
	_ = serviceManager

	// Initialize message queue event publisher
	eventPublisher, err := queue.NewEventPublisher(mqClient, zapLogger)
	if err != nil {
		log.Fatalf("Failed to create event publisher: %v", err)
	}

	// Initialize services
	userService := services.NewUserService(userRepo, jwtManager)
	orgService := services.NewOrganizationService(orgRepo, userRepo)
	templateService := services.NewTemplateService(templateRepo, orgRepo)
	volumeService := services.NewVolumeService(volumeRepo, eventPublisher, zapLogger)
	workspaceService := services.NewWorkspaceService(workspaceRepo, templateRepo, volumeRepo, eventPublisher, zapLogger)
	quotaService := services.NewQuotaService(quotaRepo, workspaceRepo, volumeRepo)
	
	// Note: quotaService will be integrated in Phase 2 for resource quota enforcement
	_ = quotaService

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	orgHandler := handlers.NewOrganizationHandler(orgService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	volumeHandler := handlers.NewVolumeHandler(volumeService)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)

	// Create and setup API server
	server := api.NewServer(gormDB, jwtManager)
	server.SetupRoutes(userHandler, orgHandler, templateHandler, workspaceHandler, volumeHandler)

	// Start API server in a goroutine
	go func() {
		addr := cfg.ServerAddress
		if addr == "" {
			addr = ":8080"
		}
		log.Infof("Starting API server on %s", addr)
		if err := server.Run(addr); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// TODO: Start background workers for workspace reconciliation
	// Create controller for background tasks
	// ctrl := controller.NewController(k8sClient, db, mqClient, log)
	// ctx, cancel := context.WithCancel(context.Background())
	// go func() {
	// 	if err := ctrl.Start(ctx); err != nil {
	// 		log.Errorf("Controller error: %v", err)
	// 	}
	// }()

	log.Info("Controller started successfully")
	// Setup context for graceful shutdown
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start controller
	// go func() {
	// 	if err := ctrl.Start(ctx); err != nil {
	// 		log.Fatalf("Controller error: %v", err)
	// 	}
	// }()

	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutting down gracefully...")
	// cancel()

	log.Info("Controller stopped")
}
