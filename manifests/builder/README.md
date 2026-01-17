# IdeKube Builder Image Documentation

## Overview

The `manifests/build/Dockerfile` builds a complete CI/CD environment image for compiling all IdeKube Docker images within containers.

## Features

✅ **Docker-in-Docker (DinD)** - Run Docker inside containers
✅ **Docker Buildx** - Support multi-platform image builds (linux/amd64, linux/arm64)
✅ **Go 1.21** - For Controller builds and Swagger generation
✅ **Node.js 18 + Yarn** - For Frontend builds
✅ **Java 11** - For OpenAPI Generator CLI
✅ **OpenAPI Generator CLI** - Auto-generate TypeScript API clients
✅ **All Required Tools** - Git, Make, curl, wget, etc.

## Building the Image

```bash
# Build from project root
docker build -f manifests/build/Dockerfile -t idekube-builder:latest .

# Or use specific tags
docker build -f manifests/build/Dockerfile -t idekube-builder:v1.0.0 .
docker build -f manifests/build/Dockerfile -t myregistry.com/idekube-builder:latest .
```

## Usage

### Basic Usage

```bash
# Build all images (without push)
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  idekube-builder:latest \
  ./scripts/build-all-docker-images.sh

# Build and push to Docker Hub
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  idekube-builder:latest \
  ./scripts/build-all-docker-images.sh --push
```

### Advanced Usage

```bash
# Specify custom registry and version
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  idekube-builder:latest \
  ./scripts/build-all-docker-images.sh \
    --registry myregistry.com \
    --version v1.0.0 \
    --platforms linux/amd64,linux/arm64 \
    --push

# Build only specific platform (for local testing)
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  idekube-builder:latest \
  ./scripts/build-all-docker-images.sh \
    --platforms linux/amd64
```

## Environment Variables

The following environment variables can be set in the container:

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  -e DOCKER_REGISTRY=myregistry.com \
  -e VERSION=v1.0.0 \
  idekube-builder:latest \
  ./scripts/build-all-docker-images.sh
```

## Using in CI/CD

### GitHub Actions

```yaml
- name: Build Docker images
  uses: docker/build-push-action@v5
  with:
    context: .
    file: manifests/build/Dockerfile
    tags: idekube-builder:latest
    push: false
    load: true

- name: Run builds in builder container
  run: |
    docker run --rm \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -v ${{ github.workspace }}:/workspace \
      idekube-builder:latest \
      ./scripts/build-all-docker-images.sh --push
```

### GitLab CI

```yaml
build-with-builder-image:
  stage: build
  image: docker:dind
  services:
    - docker:dind
  script:
    # Build builder image
    - docker build -f manifests/build/Dockerfile -t idekube-builder:latest .
    
    # Run builds
    - |
      docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v $CI_PROJECT_DIR:/workspace \
        idekube-builder:latest \
        ./scripts/build-all-docker-images.sh --push
```

## Installed Tool Versions

```
Docker:              Latest (dind)
Docker Buildx:       v0.12.0
Go:                  1.21
Node.js:             18
Yarn:                Latest
Java (OpenJDK):      11
OpenAPI Generator:   Latest
Swag (Go Swagger):   Latest
Git:                 Latest
Make:                Latest
```

## Build Process

When the image starts, the following steps are automatically executed:

1. ✅ Wait for Docker daemon to be ready
2. ✅ Configure Docker buildx builder (multi-platform support)
3. ✅ Initialize buildx build environment
4. ✅ Start Docker daemon and wait for command execution

## Performance Optimization

### Caching Strategy

- Builds are fully isolated to avoid interference with local Docker
- Supports Docker BuildKit caching
- Supports parallel multi-platform builds

### Disk Usage

```bash
# Clean build layer cache (if needed)
docker builder prune

# Complete rebuild (ignore cache)
docker build --no-cache -f manifests/build/Dockerfile -t idekube-builder:latest .
```

## Troubleshooting

### Docker Socket Permission Issues

```bash
# Ensure Docker socket is accessible
ls -la /var/run/docker.sock

# May need to adjust permissions
sudo chmod 666 /var/run/docker.sock
```

### Build Timeout

Increase Docker timeout settings:

```bash
docker run --rm \
  -e DOCKER_CLIENT_TIMEOUT=120 \
  -e COMPOSE_HTTP_TIMEOUT=120 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  idekube-builder:latest \
  ./scripts/build-all-docker-images.sh
```

### Multi-platform Build Failures

Ensure QEMU emulation is enabled:

```bash
# On host machine
docker run --privileged --rm tonistiigi/binfmt --install all
```

## Best Practices

1. **Regular Updates** - Regularly rebuild the image to get the latest dependencies
2. **Tag Management** - Use version tags to track image updates
3. **Isolated Environment** - Use this image to keep CI/CD environment clean
4. **Resource Limits** - Configure memory and CPU limits as needed

## Related Files

- [scripts/build-all-docker-images.sh](../../scripts/build-all-docker-images.sh) - Main build script
- [.github/workflows/build-docker-images.yml](../../.github/workflows/build-docker-images.yml) - GitHub Actions workflow
- [.gitlab-ci.yml](../../.gitlab-ci.yml) - GitLab CI configuration

---

**Last Updated: January 2026**
