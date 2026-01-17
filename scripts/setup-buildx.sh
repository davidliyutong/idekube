#!/usr/bin/env bash

################################################################################
# Script: setup-buildx.sh
# Purpose: Setup Docker buildx for multi-platform builds
# This script should be run before using build-all-docker-images.sh
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Docker Buildx Setup ===${NC}"
echo ""

# Check Docker installation
if ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ Docker is not installed${NC}"
    echo "Please install Docker from https://www.docker.com"
    exit 1
fi

echo -e "${GREEN}✓ Docker found:${NC} $(docker --version)"

# Check buildx
if ! docker buildx version &> /dev/null; then
    echo -e "${YELLOW}⚠ Docker buildx is not available${NC}"
    echo ""
    echo "Options to install buildx:"
    echo "  1. Update Docker Desktop to latest version"
    echo "  2. Install buildx manually:"
    echo "     brew install docker-buildx  (macOS)"
    echo "     apt-get install docker-buildx-plugin  (Linux)"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓ Docker buildx found:${NC} $(docker buildx version | head -1)"
echo ""

# List available builders
echo -e "${BLUE}Current builders:${NC}"
docker buildx ls
echo ""

# Create or use idekube builder
BUILDER_NAME="idekube-builder"

if docker buildx ls | grep -q "^${BUILDER_NAME} "; then
    echo -e "${YELLOW}⚠ Builder '${BUILDER_NAME}' already exists${NC}"
    echo "Using existing builder..."
    docker buildx use "${BUILDER_NAME}"
else
    echo -e "${BLUE}Creating new builder: ${BUILDER_NAME}${NC}"
    docker buildx create \
        --name "${BUILDER_NAME}" \
        --platform linux/amd64,linux/arm64 \
        --use
    echo -e "${GREEN}✓ Builder created${NC}"
fi

echo ""
echo -e "${BLUE}Inspecting builder:${NC}"
docker buildx inspect "${BUILDER_NAME}" || true

echo ""
echo -e "${GREEN}✓ Setup complete!${NC}"
echo ""
echo "You can now use:"
echo "  ./scripts/build-all-docker-images.sh --platforms linux/amd64,linux/arm64 --push"
echo ""
echo "Or use Makefile:"
echo "  make docker-build-multiarch-push"
