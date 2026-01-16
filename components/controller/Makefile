.PHONY: build build-migrate run test clean docker-build docker-push deps migrate-up migrate-down migrate-version help

# Variables
APP_NAME := idekube-controller
MIGRATE_TOOL := migrate
DOCKER_REGISTRY ?= docker.io
DOCKER_IMAGE := ${DOCKER_REGISTRY}/davidliyutong/${APP_NAME}
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "latest")

# Build the application
build:
	@echo "Building ${APP_NAME}..."
	go build -o bin/${APP_NAME} ./cmd/controller
	@echo "Build complete: bin/${APP_NAME}"

# Build migration CLI tool
build-migrate:
	@echo "Building migration tool..."
	go build -o bin/${MIGRATE_TOOL} ./cmd/migrate
	@echo "Build complete: bin/${MIGRATE_TOOL}"

# Run the application
run:
	go run ./cmd/controller

# Run with hot reload (requires air)
dev:
	@echo "Starting development mode with hot reload..."
	@air

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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

# Database migration commands
migrate-up: build-migrate
	@echo "Running database migrations..."
	@./bin/${MIGRATE_TOOL} -action=up

migrate-down: build-migrate
	@echo "Rolling back last migration..."
	@./bin/${MIGRATE_TOOL} -action=down

migrate-version: build-migrate
	@echo "Checking migration version..."
	@./bin/${MIGRATE_TOOL} -action=version

# Legacy migration command (deprecated)
db-migrate:
	@echo "⚠️  WARNING: db-migrate is deprecated. Use 'make migrate-up' instead."
	@echo "This command uses the old migration method and will be removed in a future version."
	@psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f migrations/001_init_schema.sql

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run || true

# Generate Swagger documentation
API_GEN_DIR := ./docs/api
swagger-gen:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/controller/main.go -o ${API_GEN_DIR} --parseDependency --parseInternal
	@echo "Swagger documentation generated in ${API_GEN_DIR}"

# Generate JavaScript client from Swagger spec
swagger-js-client:
	@echo "Generating JavaScript client from Swagger spec..."
	@which swagger-codegen > /dev/null || (echo "Error: swagger-codegen not found. Install it first: brew install swagger-codegen" && exit 1)
	@mkdir -p ${API_GEN_DIR}/client-js
	@swagger-codegen generate -i ${API_GEN_DIR}/swagger.json -l javascript -o ${API_GEN_DIR}/client-js
	@echo "JavaScript client generated in ${API_GEN_DIR}/client-js"

# Alternative: Generate TypeScript client (using openapi-generator)
swagger-ts-client:
	@echo "Generating TypeScript client from Swagger spec..."
	@which openapi-generator > /dev/null || (echo "Error: openapi-generator not found. Install it first: npm install -g @openapitools/openapi-generator-cli" && exit 1)
	@mkdir -p ${API_GEN_DIR}/client-ts
	@openapi-generator generate -i ${API_GEN_DIR}/swagger.json -g typescript-fetch -o ${API_GEN_DIR}/client-ts
	@echo "TypeScript client generated in ${API_GEN_DIR}/client-ts"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf ${API_GEN_DIR}/

# Build Docker image
docker-build:
	docker build -t ${DOCKER_IMAGE}:${VERSION} .
	docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest

# Push Docker image
docker-push: docker-build
	docker push ${DOCKER_IMAGE}:${VERSION}

# Help command
help:
	@echo "IDEKube Controller Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              - Build the controller application"
	@echo "  make build-migrate      - Build the migration CLI tool"
	@echo ""
	@echo "Run Commands:"
	@echo "  make run                - Run the controller"
	@echo "  make dev                - Run with hot reload (requires air)"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test               - Run tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo ""
	@echo "Migration Commands:"
	@echo "  make migrate-up         - Apply all pending migrations"
	@echo "  make migrate-down       - Roll back the last migration"
	@echo "  make migrate-version    - Check current migration version"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                - Format code"
	@echo "  make lint               - Lint code (requires golangci-lint)"
	@echo ""
	@echo "Swagger/API Documentation:"
	@echo "  make swagger-gen        - Generate Swagger documentation"
	@echo "  make swagger-js-client  - Generate JavaScript client from Swagger"
	@echo "  make swagger-ts-client  - Generate TypeScript client from Swagger"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-push        - Build and push Docker image"
	@echo ""
	@echo "Other Commands:"
	@echo "  make deps               - Install dependencies"
	@echo "  make deps-update        - Update dependencies"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make help               - Show this help message"
