package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Kubeconfig         string
	Namespace          string
	Postgres           PostgresConfig
	Redis              RedisConfig
	SMTP               SMTPConfig
	LogLevel           string
	WorkerThreads      int
	ServerAddress      string
	JWTSecret          string
	AdminPassword      string
	EnableRegistration bool
	BaseURL            string
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

type SMTPConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	UseTLS    bool
	FromEmail string
	FromName  string
}

func Load() (*Config, error) {
	cfg := &Config{
		Kubeconfig:    os.Getenv("KUBECONFIG"),
		Namespace:     GetEnvOrDefault("NAMESPACE", ""),
		ServerAddress: GetEnvOrDefault("SERVER_ADDRESS", ":8080"),
		JWTSecret:     GetEnvOrDefault("JWT_SECRET", "change-me-in-production"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		Postgres: PostgresConfig{
			Host:     GetEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     GetEnvAsIntOrDefault("POSTGRES_PORT", 5432),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Database: os.Getenv("POSTGRES_DB"),
		},
		Redis: RedisConfig{
			Host:     GetEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     GetEnvAsIntOrDefault("REDIS_PORT", 6379),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       GetEnvAsIntOrDefault("REDIS_DB", 0),
		},
		SMTP: SMTPConfig{
			Host:      GetEnvOrDefault("SMTP_HOST", "localhost"),
			Port:      GetEnvAsIntOrDefault("SMTP_PORT", 587),
			User:      GetEnvOrDefault("SMTP_USER", ""),
			Password:  GetEnvOrDefault("SMTP_PASSWORD", ""),
			UseTLS:    GetEnvAsBoolOrDefault("SMTP_USE_TLS", true),
			FromEmail: GetEnvOrDefault("SMTP_FROM_EMAIL", "noreply@idekube.io"),
			FromName:  GetEnvOrDefault("SMTP_FROM_NAME", "IDEKube"),
		},
		LogLevel:           GetEnvOrDefault("LOG_LEVEL", "info"),
		WorkerThreads:      GetEnvAsIntOrDefault("WORKER_THREADS", 1),
		EnableRegistration: GetEnvAsBoolOrDefault("ENABLE_REGISTRATION", true),
		BaseURL:            GetEnvOrDefault("BASE_URL", "http://localhost:8080"),
	}

	// Validate required fields
	if cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.Database == "" {
		return nil, fmt.Errorf("PostgreSQL configuration is incomplete")
	}

	return cfg, nil
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "True" || value == "TRUE" || value == "1"
	}
	return defaultValue
}
