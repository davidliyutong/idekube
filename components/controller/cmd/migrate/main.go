package main

import (
	"fmt"
	"os"

	"github.com/davidliyutong/idekube-controller/internal/config"
	"github.com/davidliyutong/idekube-controller/pkg/database"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	migrationsPath string
	zapLogger      *zap.Logger
)

func main() {
	// Initialize zap logger
	var err error
	zapLogger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "IDEKube Database Migration Tool",
		Long:  "A database migration tool for IDEKube Controller to manage schema versions",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&migrationsPath, "path", "p", "./migrations", "Path to migrations directory")

	// Add subcommands
	rootCmd.AddCommand(upCmd())
	rootCmd.AddCommand(downCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func upCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		Long:  "Apply all pending database migrations to bring the schema up to date",
		RunE: func(cmd *cobra.Command, args []string) error {
			zapLogger.Info("Running migrations up...")

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			db, err := database.NewPostgresClient(cfg.Postgres)
			if err != nil {
				return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
			}
			defer db.Close()

			migrationConfig := database.MigrationConfig{
				MigrationsPath: migrationsPath,
				DatabaseName:   cfg.Postgres.Database,
			}

			if err := database.RunMigrations(db, migrationConfig); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			zapLogger.Info("Migrations completed successfully")
			return nil
		},
	}
}

func downCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		Long:  "Revert the most recently applied database migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			zapLogger.Info("Rolling back last migration...")

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			db, err := database.NewPostgresClient(cfg.Postgres)
			if err != nil {
				return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
			}
			defer db.Close()

			migrationConfig := database.MigrationConfig{
				MigrationsPath: migrationsPath,
				DatabaseName:   cfg.Postgres.Database,
			}

			if err := database.MigrateDown(db, migrationConfig); err != nil {
				return fmt.Errorf("failed to rollback migration: %w", err)
			}

			zapLogger.Info("Migration rolled back successfully")
			return nil
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current migration version",
		Long:  "Display the current database schema version",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			db, err := database.NewPostgresClient(cfg.Postgres)
			if err != nil {
				return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
			}
			defer db.Close()

			migrationConfig := database.MigrationConfig{
				MigrationsPath: migrationsPath,
				DatabaseName:   cfg.Postgres.Database,
			}

			version, dirty, err := database.GetMigrationVersion(db, migrationConfig)
			if err != nil {
				return fmt.Errorf("failed to get migration version: %w", err)
			}

			if dirty {
				fmt.Printf("Current migration version: %d (dirty - manual intervention required)\n", version)
			} else {
				fmt.Printf("Current migration version: %d\n", version)
			}

			return nil
		},
	}
}
