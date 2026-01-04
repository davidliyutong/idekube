package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/davidliyutong/idekube-rbac/internal/config"
)

type PostgresClient struct {
	db *sql.DB
}

func NewPostgresClient(cfg config.PostgresConfig) (*PostgresClient, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresClient{db: db}, nil
}

func (c *PostgresClient) Close() error {
	return c.db.Close()
}

func (c *PostgresClient) DB() *sql.DB {
	return c.db
}

// TODO: Add database operation methods
// Example:
// func (c *PostgresClient) CreateResource(ctx context.Context, resource *Resource) error
// func (c *PostgresClient) GetResource(ctx context.Context, id string) (*Resource, error)
// func (c *PostgresClient) UpdateResource(ctx context.Context, resource *Resource) error
// func (c *PostgresClient) DeleteResource(ctx context.Context, id string) error
