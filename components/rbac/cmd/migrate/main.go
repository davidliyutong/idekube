package main

import (
	"fmt"
	"os"

	"github.com/davidliyutong/idekube-rbac/internal/config"
	"github.com/davidliyutong/idekube-rbac/pkg/database"
	"github.com/davidliyutong/idekube-rbac/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	migrationsPath string
	log            *logger.Logger
)

func main() {
	// Initialize logger
	log = logger.NewLogger()

	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "IDEKube RBAC Database Migration Tool",
		Long:  "A database migration tool for IDEKube RBAC service to manage schema versions",
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
			log.Info("Running migrations up...")
			
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
			
			log.Info("Migrations completed successfully")
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
			log.Info("Rolling back last migration...")
			
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
			
			log.Info("Migration rolled back successfully")
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
				log.Warnf("Current migration version: %d (dirty - manual intervention required)", version)
			} else {
				log.Infof("Current migration version: %d", version)
			}
			
			return nil
		},
	}
}
