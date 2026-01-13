# idekube-housekeeper

idekube-housekeeper is the cleanup and maintenance component of the idekube platform, responsible for managing the lifecycle of Kubernetes resources, cleaning up expired resources, archiving data, and executing periodic maintenance tasks.

## Overview

idekube-housekeeper receives cleanup requests by listening to RabbitMQ message queues, interacting with PostgreSQL database and Kubernetes API to perform the following tasks:

- **Resource Cleanup**: Delete expired Kubernetes workspace resources (Namespace, Deployment, Service, PVC, etc.)
- **Data Archiving**: Archive workspace data to persistent storage
- **Periodic Maintenance**: Execute database optimization, image cleanup, and other periodic maintenance tasks
- **State Synchronization**: Keep database records consistent with Kubernetes resource states

## Core Features

### 1. Workspace Cleanup
- Receive cleanup requests from RabbitMQ
- Delete workspace resources in Kubernetes
- Update workspace status in database
- Support forced cleanup and safe cleanup modes

### 2. Data Archiving
- Archive workspace data to S3/NFS storage
- Compress and package workspace files
- Set data retention periods
- Support incremental and full backups

### 3. Periodic Maintenance
- Clean up workspaces with timeout and no access
- Optimize PostgreSQL database
- Clean up unused container images
- Detect and repair resource inconsistencies

### 4. Monitoring and Logging
- Detailed operation logs
- Cleanup task statistics and reports
- Error tracking and alerting
- Performance metrics collection

## Architecture

```
┌─────────────┐
│   RabbitMQ  │  ← Receive cleanup requests
└──────┬──────┘
       │
       ↓
┌─────────────────┐
│  Housekeeper    │
│                 │
│  ┌───────────┐  │
│  │  Message   │  │
│  │ Processing │  │
│  └───────────┘  │
│        │        │
│        ↓        │
│  ┌───────────┐  │
│  │  Resource  │  │
│  │  Cleanup   │  │
│  └───────────┘  │
│        │        │
│        ↓        │
│  ┌───────────┐  │
│  │  Status    │  │
│  │  Update    │  │
│  └───────────┘  │
└────┬─────┬──────┘
     │     │
     ↓     ↓
┌─────────┐ ┌──────────┐
│Kubernetes│ │PostgreSQL│
└─────────┘ └──────────┘
```

## Quick Start

### Simplest Way - Using Docker Compose

```bash
# Clone project
git clone https://github.com/davidliyutong/idekube.git
cd idekube/components/housekeeper

# Configure environment variables (edit .env file)
cp .env.example .env

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f housekeeper
```

### Local Development

```bash
# Install dependencies
go mod download

# Build
make build

# Run (PostgreSQL and RabbitMQ must be started first)
make run
```

### Kubernetes Deployment

```bash
# Using Helm Chart (recommended)
helm upgrade --install idekube ../../helm \
  --namespace idekube-system \
  --create-namespace

# Or using kubectl
kubectl apply -f k8s/
```

For detailed deployment guide, please refer to [Quick Start Documentation](./QUICKSTART.md).

## Documentation

### 📚 Complete Documentation Navigation

- **[Quick Start Guide (QUICKSTART.md)](./docs/QUICKSTART.md)** - Deployment and configuration guide
  - System requirements
  - Local development environment setup
  - Docker deployment
  - Kubernetes deployment
  - Configuration explanation
  - Troubleshooting

- **[API Definition (API.md)](./docs/API.md)** - RabbitMQ message formats and database models
  - RabbitMQ connection configuration
  - Queue definitions and message formats
  - Database model description
  - Message sending examples (Python, Go, cURL)
  - Monitoring and metrics

- **[API Testing Guide (API_TESTING.md)](./docs/API_TESTING.md)** - Functional testing methods
  - Test environment preparation
  - Functional test cases
  - Integration testing
  - Performance testing
  - Failure recovery testing
  - Common issues troubleshooting

## Configuration Overview

Configuration via environment variables:

| Category | Environment Variable | Default | Description |
|----------|---------------------|---------|-------------|
| **Kubernetes** | `KUBECONFIG` | - | kubeconfig file path |
| | `NAMESPACE` | All | Namespace to watch |
| **PostgreSQL** | `POSTGRES_HOST` | localhost | Database host |
| | `POSTGRES_PORT` | 5432 | Database port |
| | `POSTGRES_USER` | - | Username (required) |
| | `POSTGRES_PASSWORD` | - | Password (required) |
| | `POSTGRES_DB` | - | Database name (required) |
| **RabbitMQ** | `RABBITMQ_HOST` | localhost | Message queue host |
| | `RABBITMQ_PORT` | 5672 | Message queue port |
| | `RABBITMQ_USER` | - | Username (required) |
| | `RABBITMQ_PASSWORD` | - | Password (required) |
| | `RABBITMQ_VHOST` | / | Virtual host |
| **Application** | `LOG_LEVEL` | info | Log level |
| | `WORKER_THREADS` | 1 | Number of worker threads |

For detailed configuration instructions, please refer to [Quick Start Guide](./docs/QUICKSTART.md#configuration).

## Development

### Prerequisites
- Go 1.21+
- Docker (for building images)
- Kubernetes cluster (for testing)
- PostgreSQL 12+
- RabbitMQ 3.8+

### Build

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Run tests
make test

# Clean build artifacts
make clean
```

### Project Structure

```
.
├── cmd/
│   └── housekeeper/        # Main program entry
├── internal/
│   ├── config/            # Configuration management
│   ├── housekeeper/       # Core business logic
│   └── models/            # Data models
├── pkg/
│   ├── database/          # Database client
│   ├── k8s/               # Kubernetes client
│   ├── logger/            # Logging utility
│   └── queue/             # RabbitMQ client
├── bin/                   # Compiled output directory
├── Dockerfile            # Docker image build file
├── Makefile              # Build scripts
└── README.md             # This file
```

### Development Workflow

1. **Fork and clone repository**
2. **Create feature branch**: `git checkout -b feature/your-feature`
3. **Write code and tests**
4. **Run tests**: `make test`
5. **Commit changes**: `git commit -am 'Add some feature'`
6. **Push branch**: `git push origin feature/your-feature`
7. **Create Pull Request**

## Technology Stack

- **Language**: Go 1.21+
- **Message Queue**: RabbitMQ (AMQP 0-9-1)
- **Database**: PostgreSQL 12+
- **Container Orchestration**: Kubernetes 1.20+
- **Dependency Management**: Go Modules

### Main Dependencies

- `k8s.io/client-go` - Kubernetes client
- `github.com/rabbitmq/amqp091-go` - RabbitMQ client
- `gorm.io/gorm` - ORM framework
- `github.com/lib/pq` - PostgreSQL driver

## Contributing

Contributions of code, bug reports, and improvement suggestions are welcome!

### How to Contribute

1. Check [GitHub Issues](https://github.com/davidliyutong/idekube/issues)
2. Fork the project
3. Create feature branch
4. Submit Pull Request

### Code Standards

- Follow Go coding standards and best practices
- Add tests for new features
- Update relevant documentation
- Use clear descriptions in commit messages

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details

## Related Links

- **Main Project**: [idekube](https://github.com/davidliyutong/idekube)
- **Controller Component**: [../controller](../controller)
- **Frontend Component**: [../frontend](../frontend)
- **RBAC Component**: [../rbac](../rbac)

## Getting Help

- **GitHub Issues**: [Submit Issues](https://github.com/davidliyutong/idekube/issues)
- **Documentation**: See documentation navigation above
- **Examples**: Refer to [API Testing Guide](./docs/API_TESTING.md)

---

**Maintainers**: idekube Team  
**Last Updated**: 2026-01-13
