# IdeKube CI/CD Builder Image
# 
# This image is designed to run in CI/CD environments for building
# all IdeKube Docker images with buildx support.
#
# Features:
# - Docker-in-Docker (dind) for container building
# - Docker buildx pre-configured for multi-platform builds
# - Go toolchain for swagger documentation generation
# - Node.js and Yarn for frontend builds
# - OpenAPI Generator CLI for TypeScript code generation
# - Git, Make, and other build utilities
#
# Usage:
#   docker build -t idekube-builder:latest -f manifests/build/Dockerfile .
#   docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
#     idekube-builder:latest ./scripts/build-all-docker-images.sh

FROM docker:dind

# Set environment variables
ENV GOLANG_VERSION=1.21
ENV NODE_VERSION=18
ENV JAVA_VERSION=11
ENV BUILDKIT_PROGRESS=plain
ENV DOCKER_BUILDKIT=1

# Install build dependencies
RUN set -eux; \
    apk add --no-cache \
        bash \
        git \
        make \
        curl \
        wget \
        ca-certificates \
        openssl \
        build-base \
        python3 \
        py3-pip \
        jq \
        gettext \
        && \
    echo "✓ Base build tools installed"

# Install Go
RUN set -eux; \
    apk add --no-cache go; \
    go version; \
    echo "✓ Go installed"

# Install Go swagger generation tool (swag)
RUN set -eux; \
    go install github.com/swaggo/swag/cmd/swag@latest; \
    /root/go/bin/swag --version; \
    echo "✓ Swag installed"

# Install Node.js and Yarn
RUN set -eux; \
    apk add --no-cache nodejs npm; \
    npm install -g yarn; \
    node --version; \
    yarn --version; \
    echo "✓ Node.js and Yarn installed"

# Install OpenAPI Generator CLI (requires Java)
RUN set -eux; \
    apk add --no-cache openjdk11; \
    echo "✓ OpenJDK 11 installed"

# Install OpenAPI Generator CLI
RUN set -eux; \
    npm install -g @openapitools/openapi-generator-cli; \
    openapi-generator-cli version; \
    echo "✓ OpenAPI Generator CLI installed"

# Setup docker buildx
RUN set -eux; \
    mkdir -p /root/.docker/cli-plugins; \
    wget -O /root/.docker/cli-plugins/docker-buildx \
        https://github.com/docker/buildx/releases/download/v0.12.0/buildx-v0.12.0.linux-amd64; \
    chmod +x /root/.docker/cli-plugins/docker-buildx; \
    docker buildx version; \
    echo "✓ Docker buildx installed"

# Create buildx builder instance
# Note: This will be created at runtime when docker daemon is available
RUN set -eux; \
    mkdir -p /root/.docker; \
    echo '{"features": {"containerd-snapshotter": true}}' > /root/.docker/daemon.json; \
    echo "✓ Docker configuration prepared"

# Setup entrypoint script for buildx initialization
COPY --chmod=755 manifests/builder/01-buildx-setup.sh /docker-entrypoint.d/01-buildx-setup.sh

# Set working directory
WORKDIR /workspace

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD docker ps > /dev/null 2>&1 || exit 1

# Print versions
RUN set -eux; \
    echo "=== Build Environment Summary ==="; \
    echo "Docker version:"; \
    docker --version; \
    echo ""; \
    echo "Go version:"; \
    go version; \
    echo ""; \
    echo "Node.js version:"; \
    node --version; \
    echo ""; \
    echo "Yarn version:"; \
    yarn --version; \
    echo ""; \
    echo "OpenJDK version:"; \
    java -version 2>&1 | head -1; \
    echo ""; \
    echo "OpenAPI Generator CLI version:"; \
    openapi-generator-cli version 2>/dev/null || echo "Will be initialized at runtime"; \
    echo ""; \
    echo "Make version:"; \
    make --version | head -1; \
    echo ""; \
    echo "Git version:"; \
    git --version; \
    echo "================================="

# Override the dind entrypoint to ensure buildx setup
ENTRYPOINT ["/bin/sh"]
CMD ["sh", "-c", "bash /docker-entrypoint.d/01-buildx-setup.sh && dockerd"]
