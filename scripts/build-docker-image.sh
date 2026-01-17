#!/usr/bin/env bash

################################################################################
# Script: build-docker-image.sh
# Purpose: Generic Docker image builder for CI/CD environments
# This script can be used to build images with proper caching
################################################################################

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Parse arguments
COMPONENT="$1"
REGISTRY="${2:-docker.io}"
PROJECT_NAME="davidliyutong"
VERSION="${3:-latest}"
PUSH="${4:-false}"

if [ -z "$COMPONENT" ]; then
    echo -e "${RED}Usage: $0 <component> [registry] [version] [push]${NC}"
    echo -e "${YELLOW}Components: frontend, controller, housekeeper${NC}"
    exit 1
fi

COMPONENTS_DIR="$(dirname "$SCRIPT_DIR")/components"
COMPONENT_DIR="${COMPONENTS_DIR}/${COMPONENT}"

if [ ! -d "$COMPONENT_DIR" ]; then
    echo -e "${RED}Component directory not found: $COMPONENT_DIR${NC}"
    exit 1
fi

# Component-specific configuration
case "$COMPONENT" in
    frontend)
        IMAGE_NAME="${PROJECT_NAME}/idekube-frontend"
    ;;
    controller)
        IMAGE_NAME="${PROJECT_NAME}/idekube-controller"
    ;;
    housekeeper)
        IMAGE_NAME="${PROJECT_NAME}/idekube-housekeeper"
    ;;
    *)
        echo -e "${RED}Unknown component: $COMPONENT${NC}"
        exit 1
    ;;
esac

FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}"

echo -e "${BLUE}Building Docker image: ${FULL_IMAGE}:${VERSION}${NC}"

cd "$COMPONENT_DIR"

# Build with Docker buildx for better caching
if docker buildx version > /dev/null 2>&1; then
    echo -e "${GREEN}Using docker buildx for build...${NC}"

    BUILD_ARGS="--tag ${FULL_IMAGE}:${VERSION} --tag ${FULL_IMAGE}:latest"

    if [ "$PUSH" = "true" ]; then
        BUILD_ARGS="${BUILD_ARGS} --push"
    else
        BUILD_ARGS="${BUILD_ARGS} --load"
    fi

    # Build with cache
    docker buildx build \
    --cache-from type=gha \
    --cache-to type=gha,mode=max \
    $BUILD_ARGS \
    .
else
    echo -e "${YELLOW}docker buildx not available, using regular docker build${NC}"

    docker build -t "${FULL_IMAGE}:${VERSION}" -t "${FULL_IMAGE}:latest" .

    if [ "$PUSH" = "true" ]; then
        docker push "${FULL_IMAGE}:${VERSION}"
        docker push "${FULL_IMAGE}:latest"
    fi
fi

echo -e "${GREEN}✓ Image built successfully: ${FULL_IMAGE}:${VERSION}${NC}"
