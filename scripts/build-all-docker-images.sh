#!/usr/bin/env bash

################################################################################
# Script: build-all-docker-images.sh
# Purpose: Build all Docker images for IdeKube in CI/CD environments
# Usage: ./build-all-docker-images.sh [options]
#
# This script manages dependencies and builds Docker images in correct order:
# 1. Generate Swagger documentation from controller
# 2. Generate TypeScript API client for frontend
# 3. Build multi-platform Docker images for all components using buildx
#
# Options:
#   --push              Push images to registry after building
#   --registry REGISTRY Override Docker registry (default: docker.io)
#   --version VERSION   Override version tag (default: git describe or 'latest')
#   --platforms PLAT    Comma-separated platforms (default: linux/amd64,linux/arm64)
#   --load              Load images into local Docker (only works with single platform)
#   --help              Show this help message
################################################################################

set -e  # Exit on error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPONENTS_DIR="${PROJECT_ROOT}/components"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
DOCKER_REGISTRY="${DOCKER_REGISTRY:-docker.io}"
PROJECT_NAME="davidliyutong"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'latest')}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
PUSH_IMAGES=false
LOAD_IMAGES=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --push)
            PUSH_IMAGES=true
            shift
        ;;
        --load)
            LOAD_IMAGES=true
            shift
        ;;
        --registry)
            DOCKER_REGISTRY="$2"
            shift 2
        ;;
        --version)
            VERSION="$2"
            shift 2
        ;;
        --platforms)
            PLATFORMS="$2"
            shift 2
        ;;
        --help)
            head -24 "$0" | tail -18
            exit 0
        ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
        ;;
    esac
done

################################################################################
# Helper Functions
################################################################################

log_section() {
    local title="$1"
    echo ""
    echo -e "${BLUE}===================================================${NC}"
    echo -e "${BLUE}${title}${NC}"
    echo -e "${BLUE}===================================================${NC}"
}

log_info() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

check_directory() {
    local dir="$1"
    if [ ! -d "$dir" ]; then
        log_error "Directory not found: $dir"
        exit 1
    fi
}

check_command() {
    local cmd="$1"
    if ! command -v "$cmd" &> /dev/null; then
        log_error "Required command not found: $cmd"
        return 1
    fi
}

################################################################################
# Pre-flight Checks
################################################################################

log_section "Pre-flight Checks"

# Check required commands
for cmd in docker git; do
    if check_command "$cmd"; then
        log_info "Found $cmd"
    else
        log_error "Please install $cmd"
        exit 1
    fi
done

# Check Docker buildx availability
if ! docker buildx version &> /dev/null; then
    log_error "Docker buildx is not available"
    log_warn "Installing buildx..."
    if ! command -v docker-buildx &> /dev/null; then
        log_error "Please install docker buildx: https://docs.docker.com/build/install-buildx/"
        exit 1
    fi
fi
log_info "Docker buildx is available"

# Validate load and push options
if [ "$LOAD_IMAGES" = true ] && [ "$PUSH_IMAGES" = true ]; then
    log_error "Cannot use both --load and --push options"
    exit 1
fi

if [ "$LOAD_IMAGES" = true ]; then
    if [[ "$PLATFORMS" == *","* ]]; then
        log_error "--load option only works with single platform"
        log_warn "Use --platforms=linux/amd64 or --platforms=linux/arm64 with --load"
        exit 1
    fi
fi

# Check directories exist
for dir in "${COMPONENTS_DIR}/controller" "${COMPONENTS_DIR}/frontend" "${COMPONENTS_DIR}/housekeeper"; do
    check_directory "$dir"
    log_info "Found component: $(basename "$dir")"
done

# Print configuration
echo ""
echo "Configuration:"
echo "  Project Root: ${PROJECT_ROOT}"
echo "  Docker Registry: ${DOCKER_REGISTRY}"
echo "  Version: ${VERSION}"
echo "  Platforms: ${PLATFORMS}"
echo "  Push Images: ${PUSH_IMAGES}"
echo "  Load Images: ${LOAD_IMAGES}"

################################################################################
# Step 1: Generate Swagger Documentation
################################################################################

log_section "Step 1: Generate Swagger Documentation"

cd "${COMPONENTS_DIR}/controller"

log_info "Running 'make swagger-gen' in controller..."
if make swagger-gen; then
    log_info "Swagger documentation generated successfully"
else
    log_error "Failed to generate Swagger documentation"
    exit 1
fi

SWAGGER_FILE="${COMPONENTS_DIR}/controller/docs/api/swagger.json"
if [ ! -f "$SWAGGER_FILE" ]; then
    log_error "swagger.json not found at $SWAGGER_FILE"
    exit 1
fi
log_info "Verified swagger.json exists: $SWAGGER_FILE"

################################################################################
# Step 2: Install Frontend Dependencies (if not already done)
################################################################################

log_section "Step 2: Prepare Frontend"

cd "${COMPONENTS_DIR}/frontend"

log_info "Checking/installing frontend dependencies..."
if [ ! -d "node_modules" ]; then
    log_warn "node_modules not found, running 'make install'..."
    if ! make install; then
        log_error "Failed to install frontend dependencies"
        exit 1
    fi
    log_info "Frontend dependencies installed"
else
    log_info "node_modules already exists, skipping install"
fi

################################################################################
# Step 3: Generate TypeScript API Client
################################################################################

log_section "Step 3: Generate TypeScript API Client"

cd "${COMPONENTS_DIR}/frontend"

log_info "Running 'make generate-api' in frontend..."
if make generate-api; then
    log_info "TypeScript API client generated successfully"
else
    log_error "Failed to generate TypeScript API client"
    exit 1
fi

################################################################################
# Step 4: Build Frontend Docker Image
################################################################################

log_section "Step 4: Build Frontend Docker Image"

cd "${COMPONENTS_DIR}/frontend"

FRONTEND_IMAGE="${DOCKER_REGISTRY}/${PROJECT_NAME}/idekube-frontend"

log_info "Building multi-platform Docker image: ${FRONTEND_IMAGE}:${VERSION}"
log_info "Platforms: ${PLATFORMS}"

BUILD_FLAGS="--tag ${FRONTEND_IMAGE}:${VERSION} --tag ${FRONTEND_IMAGE}:latest"

# Add load or push flags
if [ "$LOAD_IMAGES" = true ]; then
    BUILD_FLAGS="${BUILD_FLAGS} --load"
    elif [ "$PUSH_IMAGES" = true ]; then
    BUILD_FLAGS="${BUILD_FLAGS} --push"
fi

if docker buildx build --platform="${PLATFORMS}" ${BUILD_FLAGS} .; then
    log_info "Frontend Docker image built: ${FRONTEND_IMAGE}:${VERSION}"
else
    log_error "Failed to build frontend Docker image"
    exit 1
fi

################################################################################
# Step 5: Build Controller Docker Image
################################################################################

log_section "Step 5: Build Controller Docker Image"

cd "${COMPONENTS_DIR}/controller"

CONTROLLER_IMAGE="${DOCKER_REGISTRY}/${PROJECT_NAME}/idekube-controller"

log_info "Building multi-platform Docker image: ${CONTROLLER_IMAGE}:${VERSION}"
log_info "Platforms: ${PLATFORMS}"

BUILD_FLAGS="--tag ${CONTROLLER_IMAGE}:${VERSION} --tag ${CONTROLLER_IMAGE}:latest"

if [ "$LOAD_IMAGES" = true ]; then
    BUILD_FLAGS="${BUILD_FLAGS} --load"
    elif [ "$PUSH_IMAGES" = true ]; then
    BUILD_FLAGS="${BUILD_FLAGS} --push"
fi

if docker buildx build --platform="${PLATFORMS}" ${BUILD_FLAGS} .; then
    log_info "Controller Docker image built: ${CONTROLLER_IMAGE}:${VERSION}"
else
    log_error "Failed to build controller Docker image"
    exit 1
fi

################################################################################
# Step 6: Build Housekeeper Docker Image
################################################################################

log_section "Step 6: Build Housekeeper Docker Image"

cd "${COMPONENTS_DIR}/housekeeper"

HOUSEKEEPER_IMAGE="${DOCKER_REGISTRY}/${PROJECT_NAME}/idekube-housekeeper"

log_info "Building multi-platform Docker image: ${HOUSEKEEPER_IMAGE}:${VERSION}"
log_info "Platforms: ${PLATFORMS}"

BUILD_FLAGS="--tag ${HOUSEKEEPER_IMAGE}:${VERSION} --tag ${HOUSEKEEPER_IMAGE}:latest"

if [ "$LOAD_IMAGES" = true ]; then
    BUILD_FLAGS="${BUILD_FLAGS} --load"
    elif [ "$PUSH_IMAGES" = true ]; then
    BUILD_FLAGS="${BUILD_FLAGS} --push"
fi

if docker buildx build --platform="${PLATFORMS}" ${BUILD_FLAGS} .; then
    log_info "Housekeeper Docker image built: ${HOUSEKEEPER_IMAGE}:${VERSION}"
else
    log_error "Failed to build housekeeper Docker image"
    exit 1
fi

################################################################################
# Step 7: Push Images (Optional)
################################################################################

if [ "$PUSH_IMAGES" = true ]; then
    log_section "Step 7: Images Pushed to Registry"
    log_info "Images have been pushed during build process"
    elif [ "$LOAD_IMAGES" = true ]; then
    log_section "Step 7: Images Loaded Locally"
    log_info "Images have been loaded into local Docker daemon"
fi

################################################################################
# Summary
################################################################################

log_section "Build Summary"

echo "All Docker images built successfully!"
echo ""
echo "Built Images:"
echo "  • ${FRONTEND_IMAGE}:${VERSION}"
echo "  • ${CONTROLLER_IMAGE}:${VERSION}"
echo "  • ${HOUSEKEEPER_IMAGE}:${VERSION}"
echo ""
echo "Configuration:"
echo "  Platforms: ${PLATFORMS}"

if [ "$PUSH_IMAGES" = true ]; then
    echo "  Status: Pushed to ${DOCKER_REGISTRY}"
    elif [ "$LOAD_IMAGES" = true ]; then
    echo "  Status: Loaded into local Docker"
else
    echo "  Status: Built locally (use --push to push to registry, or --load for single platform)"
fi

echo ""
log_info "Build completed successfully!"
