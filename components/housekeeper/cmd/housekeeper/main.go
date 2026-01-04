package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/davidliyutong/idekube-housekeeper/internal/config"
	"github.com/davidliyutong/idekube-housekeeper/internal/housekeeper"
	"github.com/davidliyutong/idekube-housekeeper/pkg/database"
	"github.com/davidliyutong/idekube-housekeeper/pkg/k8s"
	"github.com/davidliyutong/idekube-housekeeper/pkg/logger"
	"github.com/davidliyutong/idekube-housekeeper/pkg/queue"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting idekube-housekeeper")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient(cfg.Kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Initialize PostgreSQL connection
	db, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Initialize RabbitMQ connection
	mqClient, err := queue.NewRabbitMQClient(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer mqClient.Close()

	// Create housekeeper
	hk := housekeeper.NewHousekeeper(k8sClient, db, mqClient, log)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start housekeeper
	go func() {
		if err := hk.Start(ctx); err != nil {
			log.Fatalf("Housekeeper error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutting down gracefully...")
	cancel()

	log.Info("Housekeeper stopped")
}
