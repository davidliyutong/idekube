package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/davidliyutong/idekube-rbac/internal/config"
	"github.com/davidliyutong/idekube-rbac/internal/rbac"
	"github.com/davidliyutong/idekube-rbac/pkg/database"
	"github.com/davidliyutong/idekube-rbac/pkg/k8s"
	"github.com/davidliyutong/idekube-rbac/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	kubeconfig     string
	skipMigrations bool
	migrationsPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "idekube-rbac",
		Short: "IDEKube RBAC Service",
		Long:  "IDEKube RBAC Service manages Kubernetes RBAC resources for the cloud IDE platform",
		RunE:  runRBAC,
	}

	// Add flags
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (empty for in-cluster config)")
	rootCmd.Flags().BoolVar(&skipMigrations, "skip-migrations", false, "Skip database migrations on startup")
	rootCmd.Flags().StringVar(&migrationsPath, "migrations-path", "./migrations", "Path to database migrations")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runRBAC(cmd *cobra.Command, args []string) error {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting idekube-rbac")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override kubeconfig if provided via flag
	if kubeconfig != "" {
		cfg.Kubeconfig = kubeconfig
	}

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient(cfg.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Initialize PostgreSQL connection
	db, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// Run database migrations unless skipped
	if !skipMigrations {
		log.Info("Running database migrations...")
		migrationConfig := database.MigrationConfig{
			MigrationsPath: migrationsPath,
			DatabaseName:   cfg.Postgres.Database,
		}
		if err := database.RunMigrations(db, migrationConfig); err != nil {
			return fmt.Errorf("failed to run database migrations: %w", err)
		}
		log.Info("Database migrations completed successfully")
	} else {
		log.Info("Skipping database migrations")
	}

	// Create RBAC service
	rbacService, err := rbac.NewRBACService(cfg, k8sClient, db, log)
	if err != nil {
		return fmt.Errorf("failed to initialize RBAC service: %w", err)
	}

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start RBAC service in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := rbacService.Start(ctx); err != nil {
			errChan <- fmt.Errorf("RBAC service error: %w", err)
		}
	}()

	log.Info("RBAC service started successfully")

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Info("Shutting down gracefully...")
	case err := <-errChan:
		log.Errorf("Service error: %v", err)
		return err
	}

	cancel()
	log.Info("RBAC service stopped")
	return nil
}
