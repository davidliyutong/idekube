package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Kubeconfig    string
	Namespace     string
	Postgres      PostgresConfig
	LogLevel      string
	WorkerThreads int
	HTTPPort      int
	OPA           OPAConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type OPAConfig struct {
	PolicyPath string
	DataPath   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Kubeconfig: os.Getenv("KUBECONFIG"),
		Namespace:  getEnvOrDefault("NAMESPACE", ""),
		Postgres: PostgresConfig{
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("POSTGRES_PORT", 5432),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Database: os.Getenv("POSTGRES_DB"),
		},
		LogLevel:      getEnvOrDefault("LOG_LEVEL", "info"),
		WorkerThreads: getEnvAsIntOrDefault("WORKER_THREADS", 1),
		HTTPPort:      getEnvAsIntOrDefault("HTTP_PORT", 8080),
		OPA: OPAConfig{
			PolicyPath: getEnvOrDefault("OPA_POLICY", "configs/policy.rego"),
			DataPath:   getEnvOrDefault("OPA_DATA", "configs/data.json"),
		},
	}

	// Validate required fields
	if cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.Database == "" {
		return nil, fmt.Errorf("PostgreSQL configuration is incomplete")
	}

	if cfg.OPA.PolicyPath == "" {
		return nil, fmt.Errorf("OPA configuration is incomplete")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
