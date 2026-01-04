package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Kubeconfig   string
	Namespace    string
	Postgres     PostgresConfig
	RabbitMQ     RabbitMQConfig
	LogLevel     string
	WorkerThreads int
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
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
		RabbitMQ: RabbitMQConfig{
			Host:     getEnvOrDefault("RABBITMQ_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("RABBITMQ_PORT", 5672),
			User:     os.Getenv("RABBITMQ_USER"),
			Password: os.Getenv("RABBITMQ_PASSWORD"),
			VHost:    getEnvOrDefault("RABBITMQ_VHOST", "/"),
		},
		LogLevel:      getEnvOrDefault("LOG_LEVEL", "info"),
		WorkerThreads: getEnvAsIntOrDefault("WORKER_THREADS", 1),
	}

	// Validate required fields
	if cfg.Postgres.User == "" || cfg.Postgres.Password == "" || cfg.Postgres.Database == "" {
		return nil, fmt.Errorf("PostgreSQL configuration is incomplete")
	}

	if cfg.RabbitMQ.User == "" || cfg.RabbitMQ.Password == "" {
		return nil, fmt.Errorf("RabbitMQ configuration is incomplete")
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
