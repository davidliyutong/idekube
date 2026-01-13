````markdown
# IDEKube Controller

IDEKube Controller is the core API server for a cloud IDE platform, providing workspace, template, user, and organization management functionality.

## Features

- 🔐 **User Authentication** - JWT token authentication, user registration and login
- 👥 **Organization Management** - Multi-tenant organizations, member role management
- 📦 **Template System** - Predefined workspace templates, custom resource configuration
- 💻 **Workspace Management** - Create, start, stop, and delete workspace instances
- 💾 **Volume Management** - Persistent storage, dynamic volume mounting
- 🔄 **Event-Driven** - RabbitMQ message queue integration
- 📊 **Resource Quotas** - CPU, memory, and storage quota management
- 🔍 **Audit Logs** - Complete operation audit tracking
- 📖 **Swagger Documentation** - Interactive API documentation

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Kubernetes 1.24+ (optional)

### Local Development

```bash
# 1. Install dependencies
go mod download

# 2. Configure environment variables
cp .env.example .env
# Edit .env to configure database and other parameters

# 3. Start service (automatically runs database migrations)
make run
```

After service starts, access:
- API: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/swagger/index.html
- Health check: http://localhost:8080/health

> 📖 For detailed deployment steps, refer to [Quick Start Guide](./docs/QUICKSTART.md)

## Documentation Navigation

| Document | Description |
|----------|-------------|
| [QUICKSTART.md](./docs/QUICKSTART.md) | Quick Start - Installation, configuration, and deployment guide |
| [API.md](./docs/API.md) | API Documentation - Complete interface specifications and data models |
| [API_TESTING.md](./docs/API_TESTING.md) | API Testing - Detailed testing examples and use cases |
| [DESIGN.md](./DESIGN.md) | Design Documentation - Architecture design and technical solutions |
| [Swagger UI](http://localhost:8080/swagger/index.html) | Interactive API Documentation (requires service running) |

## API Overview

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### User Management
- `GET /api/v1/users/me` - Get current user
- `POST /api/v1/users/me/password` - Change password
- `GET /api/v1/users` - List users (Admin)

### Organization Management
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations` - List organizations
- `POST /api/v1/organizations/:id/members` - Add member

### Template Management
- `POST /api/v1/templates` - Create template
- `GET /api/v1/templates` - List templates
- `GET /api/v1/templates/:id` - Get template details

### Workspace Management
- `POST /api/v1/workspaces` - Create workspace
- `GET /api/v1/workspaces` - List workspaces
- `POST /api/v1/workspaces/:id/start` - Start workspace
- `POST /api/v1/workspaces/:id/stop` - Stop workspace

### Volume Management
- `POST /api/v1/volumes` - Create volume
- `GET /api/v1/volumes` - List volumes
- `DELETE /api/v1/volumes/:id` - Delete volume

> For complete API list, check [API Documentation](./docs/API.md) or [Swagger UI](http://localhost:8080/swagger/index.html)

## Configuration

Configure the service via environment variables:

### Basic Configuration

```bash
SERVER_ADDRESS=:8080                    # Service listening address
JWT_SECRET=your-secret-key              # JWT secret
JWT_EXPIRATION_HOURS=720                # Token expiration time (hours)
```

### PostgreSQL Configuration

```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=idekube
POSTGRES_PASSWORD=idekube123
POSTGRES_DB=idekube
```

### Kubernetes Configuration (Optional)

```bash
KUBECONFIG=~/.kube/config               # kubeconfig path
NAMESPACE=idekube                       # Workspace namespace
```

### RabbitMQ Configuration (Optional)

```bash
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/
```

## Development

### Makefile Commands

```bash
# Build
make build              # Compile application
make build-migrate      # Compile migration tool

# Run
make run                # Run service
make dev                # Hot reload mode (requires air)

# Test
make test               # Run tests
make test-coverage      # Generate coverage report

# Code Quality
make fmt                # Format code
make lint               # Lint code

# Swagger Documentation
make swagger-gen        # Generate Swagger documentation
make swagger-js-client  # Generate JavaScript client
make swagger-ts-client  # Generate TypeScript client

# Database Migration
make migrate-up         # Apply migrations
make migrate-down       # Rollback migration
make migrate-version    # Check version

# Docker
make docker-build       # Build image
make docker-push        # Push image

# Clean
make clean              # Clean build artifacts
```

### Generate Swagger Documentation

```bash
# 1. Install swag tool
go install github.com/swaggo/swag/cmd/swag@latest

# 2. Generate documentation
make swagger-gen

# 3. Restart service to view
make run
# Visit http://localhost:8080/swagger/index.html
```

### Generate API Clients

```bash
# JavaScript client (requires swagger-codegen)
brew install swagger-codegen
make swagger-js-client

# TypeScript client (requires openapi-generator)
npm install -g @openapitools/openapi-generator-cli
make swagger-ts-client
```

## Deployment

### Docker

```bash
# Build image
make docker-build

# Using Docker Compose
docker-compose up -d
```

### Kubernetes

```bash
# Using Helm
helm install idekube-controller ../../helm

# Using kubectl
kubectl apply -f manifests/
```

### Database Migration

Database migrations run **automatically** on service startup; manual execution is not required. For manual control:

```bash
# Check current version
make migrate-version

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down
```

## Architecture

```
┌─────────────────┐
│   Frontend UI   │
└────────┬────────┘
         │ HTTP/REST
         ▼
┌─────────────────┐      ┌──────────────┐
│   Controller    │◄─────┤  PostgreSQL  │
│   (API Server)  │      └──────────────┘
└────────┬────────┘
         │ Events
         ▼
┌─────────────────┐      ┌──────────────┐
│   RabbitMQ      │─────►│ Housekeeper  │
└─────────────────┘      └──────┬───────┘
                                 │
                                 ▼
                         ┌──────────────┐
                         │  Kubernetes  │
                         └──────────────┘
```

- **Controller**: REST API server handling user requests
- **PostgreSQL**: Stores users, organizations, templates, workspaces, and other data
- **RabbitMQ**: Message queue for asynchronous event processing
- **Housekeeper**: Background service managing Kubernetes resources

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP Router)
- **Database**: PostgreSQL + GORM
- **Message Queue**: RabbitMQ
- **Container Orchestration**: Kubernetes
- **Authentication**: JWT
- **Documentation**: Swagger/OpenAPI

## Contributing

Contributions are welcome! Please check the [Contributing Guide](../../CONTRIBUTING.md).

## License

Apache 2.0 - See [LICENSE](../../LICENSE) file for details.

## Links

- Project Homepage: https://github.com/davidliyutong/idekube
- Issue Tracker: https://github.com/davidliyutong/idekube/issues
- Documentation: https://github.com/davidliyutong/idekube/wiki

````