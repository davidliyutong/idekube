.PHONY: install dev build preview lint type-check clean generate-api docker-build docker-run docker-dev help

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "latest")
IMAGE_NAME = davidliyutong/idekube-frontend
CONTAINER_NAME = idekube-frontend
SWAGGER_JSON = ../controller/docs/api/swagger.json

# Default target
help:
	@echo "IdeKube Frontend - Available targets:"
	@echo "  install      - Install dependencies with yarn"
	@echo "  generate-api - Generate TypeScript API client from swagger.json"
	@echo "  dev          - Run development server"
	@echo "  build        - Build for production"
	@echo "  preview      - Preview production build"
	@echo "  lint         - Lint code"
	@echo "  type-check   - Run TypeScript type checking"
	@echo "  clean        - Clean build artifacts and dependencies"
	@echo "  docker-build - Build production Docker image"
	@echo "  docker-dev   - Build and run development Docker container"
	@echo "  docker-run   - Build and run production Docker container"

# Install dependencies
install:
	yarn install

# Run development server
dev:
	yarn dev

# Build for production
build:
	yarn build

# Preview production build
preview:
	yarn preview

# Lint code
lint:
	yarn lint

# Type check
type-check:
	yarn type-check

# Generate API client from swagger.json
generate-api:
	@./scripts/generate-api.sh

# Clean build artifacts
clean:
	rm -rf dist node_modules .vite src/api/client

# Build Docker image for production
docker-build:
	docker build -t ${IMAGE_NAME}:${VERSION} .
	docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest

# Build and run development Docker container
docker-dev:
	docker build -t ${IMAGE_NAME}:dev -f Dockerfile.dev .
	docker run -it --rm \
		-p 5173:5173 \
		-v ${PWD}/src:/app/src \
		--name ${CONTAINER_NAME}-dev \
		${IMAGE_NAME}:dev

# Run Docker container for production
docker-run: docker-build
	docker run -d \
		-p 8080:80 \
		--name ${CONTAINER_NAME} \
		${IMAGE_NAME}:latest

# Stop and remove running container
docker-stop:
	docker stop ${CONTAINER_NAME} || true
	docker rm ${CONTAINER_NAME} || true

# Push to registry
docker-push: docker-build
	docker push ${IMAGE_NAME}:${VERSION}
	docker push ${IMAGE_NAME}:latest
