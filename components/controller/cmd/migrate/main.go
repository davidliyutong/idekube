package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davidliyutong/idekube-controller/internal/config"
	"github.com/davidliyutong/idekube-controller/pkg/database"
	"github.com/davidliyutong/idekube-controller/pkg/logger"
)

func main() {
	// Define command-line flags
	action := flag.String("action", "up", "Migration action: up, down, version")
	migrationsPath := flag.String("path", "./migrations", "Path to migrations directory")
	flag.Parse()

	// Initialize logger
	log := logger.NewLogger()
	log.Info("IDEKube Database Migration Tool")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize PostgreSQL connection
	db, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Migration config
	migrationConfig := database.MigrationConfig{
		MigrationsPath: *migrationsPath,
		DatabaseName:   cfg.Postgres.Database,
	}

	// Execute requested action
	switch *action {
	case "up":
		log.Info("Running migrations up...")
		if err := database.RunMigrations(db, migrationConfig); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Info("Migrations completed successfully")

	case "down":
		log.Info("Rolling back last migration...")
		if err := database.MigrateDown(db, migrationConfig); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		log.Info("Migration rolled back successfully")

	case "version":
		version, dirty, err := database.GetMigrationVersion(db, migrationConfig)
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		if dirty {
			fmt.Printf("Current migration version: %d (dirty - manual intervention required)\n", version)
		} else {
			fmt.Printf("Current migration version: %d\n", version)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", *action)
		fmt.Fprintf(os.Stderr, "Valid actions: up, down, version\n")
		os.Exit(1)
	}
}
