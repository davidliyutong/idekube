.PHONY: build run test clean docker-build docker-push

# Variables
APP_NAME := idekube-housekeeper
DOCKER_REGISTRY ?= docker.io
DOCKER_IMAGE := ${DOCKER_REGISTRY}/davidliyutong/${APP_NAME}
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "latest")

# Build the application
build:
	@echo "Building ${APP_NAME}..."
	go build -o bin/${APP_NAME} ./cmd/housekeeper
	@echo "Build complete: bin/${APP_NAME}"

# Run the application
run:
	go run ./cmd/housekeeper

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
