.PHONY: build run test clean docker-build docker-push

# Variables
APP_NAME := idekube-rbac
DOCKER_REGISTRY ?= docker.io
DOCKER_IMAGE := ${DOCKER_REGISTRY}/davidliyutong/${APP_NAME}
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "latest")

# Build the application
build:
	@echo "Building ${APP_NAME}..."
	go build -o bin/${APP_NAME} ./cmd/rbac
	@echo "Build complete: bin/${APP_NAME}"

# Build migrate tool
build-migrate:
	@echo "Building migrate tool..."
	go build -o bin/${APP_NAME}-migrate ./cmd/migrate
	@echo "Build complete: bin/${APP_NAME}-migrate"

# Build all binaries
build-all: build build-migrate

# Run the application
run:
	go run ./cmd/rbac

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t ${DOCKER_IMAGE}:${VERSION} .
	docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest

# Push Docker image
docker-push: docker-build
	docker push ${DOCKER_IMAGE}:${VERSION}

# Run with hot reload (requires air)
dev:
	@echo "Starting development mode with hot reload..."
	@air

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run || true

API_GEN_DIR := ./docs/api

# Generate swagger documentation
swagger-gen:
	@echo "Generating swagger documentation..."
	@which swag > /dev/null || (echo "swag not found. Installing..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g internal/api/server.go -o ${API_GEN_DIR} --parseDependency --parseInternal
	@echo "Swagger docs generated in ${API_GEN_DIR}"

# Generate golang client for controller
client-gen:
	@echo "Generating golang client..."
	@mkdir -p ${API_GEN_DIR}/client
	@which oapi-codegen > /dev/null || (echo "oapi-codegen not found. Installing..." && go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest)
	@echo "Converting swagger.json to OpenAPI 3.0..."
	@oapi-codegen -package client -generate types,client -o ${API_GEN_DIR}/client/client.go ${API_GEN_DIR}/swagger.json
	@echo "Golang client generated in ${API_GEN_DIR}/client/"

# View swagger docs (starts local server)
swagger-serve:
	@echo "Serving swagger docs at http://localhost:8080/swagger/"
	@echo "Make sure to build and run the application first with: make run"

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	@go run ./cmd/migrate -action up -path ./migrations

migrate-down:
	@echo "Rolling back last migration..."
	@go run ./cmd/migrate -action down -path ./migrations

migrate-version:
	@echo "Checking migration version..."
	@go run ./cmd/migrate -action version -path ./migrations

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the RBAC service binary"
	@echo "  build-migrate   - Build the migration tool binary"
	@echo "  build-all       - Build all binaries"
	@echo "  run             - Run the RBAC service"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-push     - Build and push Docker image"
	@echo "  dev             - Run with hot reload (requires air)"
	@echo "  deps            - Download dependencies"
	@echo "  deps-update     - Update dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code (requires golangci-lint)"
	@echo "  swagger-gen     - Generate swagger documentation"
	@echo "  migrate-up      - Run database migrations"
	@echo "  migrate-down    - Rollback last migration"
	@echo "  migrate-version - Check migration version"
	@echo "  help            - Show this help message"
