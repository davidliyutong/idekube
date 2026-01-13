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
# Generate swagger documentation
swagger-gen:
	@echo "Generating swagger documentation..."
	@which swag > /dev/null || (echo "swag not found. Installing..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g internal/api/server.go -o docs --parseDependency --parseInternal
	@echo "Swagger docs generated in docs/"

# Generate golang client for controller
client-gen:
	@echo "Generating golang client..."
	@mkdir -p client
	@which oapi-codegen > /dev/null || (echo "oapi-codegen not found. Installing..." && go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest)
	@echo "Converting swagger.json to OpenAPI 3.0..."
	@oapi-codegen -package client -generate types,client -o client/client.go docs/swagger.json
	@echo "Golang client generated in client/"

# View swagger docs (starts local server)
swagger-serve:
	@echo "Serving swagger docs at http://localhost:8080/swagger/"
	@echo "Make sure to build and run the application first with: make run"