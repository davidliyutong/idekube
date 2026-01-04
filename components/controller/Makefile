.PHONY: build run test clean docker-build docker-push

# Variables
APP_NAME := idekube-controller
DOCKER_REGISTRY ?= docker.io
DOCKER_IMAGE := ${DOCKER_REGISTRY}/davidliyutong/${APP_NAME}
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "latest")

# Build the application
build:
	go build -o bin/${APP_NAME} ./cmd/controller

# Run the application
run:
	go run ./cmd/controller

# Run tests
test:
	go test -v ./...

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
	air

# Install dependencies
deps:
	go mod download
	go mod tidy
