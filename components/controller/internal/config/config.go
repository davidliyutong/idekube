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
	Redis         RedisConfig
	OPA           OPAConfig
	LogLevel      string
	WorkerThreads int
	ServerAddress string
	JWTSecret     string
	AdminPassword string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type OPAConfig struct {
	PolicyPath string
	DataPath   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Kubeconfig:    os.Getenv("KUBECONFIG"),
		Namespace:     getEnvOrDefault("NAMESPACE", ""),
		ServerAddress: getEnvOrDefault("SERVER_ADDRESS", ":8080"),
		JWTSecret:     getEnvOrDefault("JWT_SECRET", "change-me-in-production"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		Postgres: PostgresConfig{
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("POSTGRES_PORT", 5432),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Database: os.Getenv("POSTGRES_DB"),
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("REDIS_PORT", 6379),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvAsIntOrDefault("REDIS_DB", 0),
		},
		OPA: OPAConfig{
			PolicyPath: getEnvOrDefault("OPA_POLICY_PATH", "configs/policy.rego"),
			DataPath:   getEnvOrDefault("OPA_DATA_PATH", "configs/data.json"),
		},
		LogLevel:      getEnvOrDefault("LOG_LEVEL", "info"),
		WorkerThreads: getEnvAsIntOrDefault("WORKER_THREADS", 1),
	}

	// Validate required fields
	if cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.Database == "" {
		return nil, fmt.Errorf("PostgreSQL configuration is incomplete")
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
