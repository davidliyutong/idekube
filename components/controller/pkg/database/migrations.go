package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	MigrationsPath string
	DatabaseName   string
}

// RunMigrations runs all pending database migrations
func RunMigrations(db *PostgresClient, config MigrationConfig) error {
	sqlDB, err := db.DB().DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		DatabaseName: config.DatabaseName,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(db *PostgresClient, config MigrationConfig) error {
	sqlDB, err := db.DB().DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		DatabaseName: config.DatabaseName,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(db *PostgresClient, config MigrationConfig) (uint, bool, error) {
	sqlDB, err := db.DB().DB()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get database instance: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		DatabaseName: config.DatabaseName,
	})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
