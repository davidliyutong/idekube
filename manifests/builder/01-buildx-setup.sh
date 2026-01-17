#!/bin/bash
set -e

echo "Waiting for Docker daemon..."
timeout 30 sh -c 'until docker ps >/dev/null 2>&1; do sleep 1; done'
echo "✓ Docker daemon is ready"

echo "Setting up buildx builder..."
# Remove existing builder if present
docker buildx ls | grep -q "^idekube-builder " && docker buildx rm idekube-builder || true

# Create new builder with multi-platform support
docker buildx create \
    --name idekube-builder \
    --platform linux/amd64,linux/arm64 \
    --use \
    2>/dev/null || true

# Use the builder
docker buildx use idekube-builder || true

echo "✓ buildx builder configured for platforms: linux/amd64,linux/arm64"
docker buildx ls
echo ""
echo "Environment ready for IdeKube builds!"
EOF