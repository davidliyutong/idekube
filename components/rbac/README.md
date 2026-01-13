# idekube-rbac

Role-Based Access Control (RBAC) service for managing permission control in the idekube platform.

## Features

- ✅ Kubernetes CRD operation support
- ✅ PostgreSQL integration
- ✅ RabbitMQ message queue integration
- ✅ Flexible permission engine based on Casbin
- ✅ RESTful API
- ✅ Swagger UI documentation
- ✅ Cloud-native configuration (environment variables)
- ✅ Golang client generation

## Quick Start

For detailed deployment guide, see [Quick Start Documentation](QUICKSTART.md).

### Local Running

```bash
# Install dependencies
make deps

# Configure environment variables (refer to QUICKSTART.md)
source .env

# Build
make build

# Run
make run
```

The service will start at `http://localhost:8080`.

Access Swagger UI: `http://localhost:8080/swagger/`

## Documentation

- **[API Definition Documentation](API.md)** - Complete API specification, data models, and integration guide
- **[API Testing Guide](API_TESTING.md)** - API testing methods and examples
- **[Quick Start](QUICKSTART.md)** - Detailed deployment and configuration guide

## API Overview

### Health Check
```bash
GET /healthz
```

### Permission Check
```bash
POST /api/v1/rbac/check
{
  "user_id": 123,
  "resource_type": "workspace",
  "resource_id": "ws-001",
  "action": "read"
}
```

### Role Assignment
```bash
POST /api/v1/rbac/assign-role
{
  "user_id": 123,
  "role": "admin"
}
```

For complete API documentation, see [API.md](API.md).

## Development

### Prerequisites

- Go 1.21+
- Docker (for building images)
- Kubernetes cluster (for testing)
- PostgreSQL 12+
- RabbitMQ 3.9+

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Push Docker image
make docker-push
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Generate Documentation and Client

```bash
# Generate Swagger documentation
make swagger-gen

# Generate Golang client (for Controller use)
make client-gen

# View Swagger UI (need to run service first)
make swagger-serve
```

Generated files:
- Swagger documentation: `docs/swagger.json`, `docs/swagger.yaml`
- Golang client: `client/client.go`

### Code Formatting and Linting

```bash
# Format code
make fmt

# Lint code
make lint
```

## Configuration

The service is configured through environment variables. For detailed configuration instructions, refer to [QUICKSTART.md](QUICKSTART.md#configuration).

### Core Configuration

#### Kubernetes
- `KUBECONFIG`: kubeconfig file path (optional when running inside cluster)
- `NAMESPACE`: namespace to watch (default: all namespaces)

#### PostgreSQL
- `POSTGRES_HOST`: PostgreSQL host (default: localhost)
- `POSTGRES_PORT`: PostgreSQL port (default: 5432)
- `POSTGRES_USER`: PostgreSQL user
- `POSTGRES_PASSWORD`: PostgreSQL password
- `POSTGRES_DB`: PostgreSQL database name

#### RabbitMQ
- `RABBITMQ_HOST`: RabbitMQ host (default: localhost)
- `RABBITMQ_PORT`: RabbitMQ port (default: 5672)
- `RABBITMQ_USER`: RabbitMQ user
- `RABBITMQ_PASSWORD`: RabbitMQ password
- `RABBITMQ_VHOST`: RabbitMQ virtual host (default: /)

#### Application Configuration
- `HTTP_PORT`: HTTP service port (default: 8080)
- `LOG_LEVEL`: Log level (default: info)
- `WORKER_THREADS`: Number of worker threads (default: 1)
- `CASBIN_MODEL_PATH`: Casbin model file path
- `CASBIN_POLICY_PATH`: Casbin policy file path

## Deployment

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e POSTGRES_HOST=postgres \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=idekube_rbac \
  -e RABBITMQ_HOST=rabbitmq \
  -e RABBITMQ_PORT=5672 \
  -e RABBITMQ_USER=guest \
  -e RABBITMQ_PASSWORD=guest \
  davidliyutong/idekube-rbac:latest
```

### Kubernetes

Deploy using the Helm Chart from the project root directory:

```bash
helm install idekube-rbac ../../helm \
  -n idekube \
  --create-namespace
```

For detailed deployment instructions, refer to [QUICKSTART.md](QUICKSTART.md#kubernetes-deployment).

## Project Structure

```
.
├── cmd/
│   └── rbac/           # Main program entry
├── internal/
│   ├── api/            # HTTP API server
│   ├── config/         # Configuration management
│   ├── permission/     # Permission check service
│   └── rbac/           # RBAC core service
├── pkg/
│   ├── database/       # Database client
│   ├── k8s/            # Kubernetes client
│   ├── logger/         # Logging utility
│   └── queue/          # Message queue client
├── configs/            # Configuration files
├── docs/               # Swagger documentation (auto-generated)
├── client/             # Golang client (auto-generated)
├── Dockerfile
├── Makefile
└── README.md
```

## Tech Stack

- **Programming Language:** Go 1.21+
- **Permission Engine:** [Casbin](https://casbin.org/) - Powerful permission management framework
- **Database:** PostgreSQL - Store Casbin policies
- **Message Queue:** RabbitMQ - Asynchronous message processing
- **API Documentation:** Swagger/OpenAPI - Auto-generate API documentation
- **Containerization:** Docker - Container deployment
- **Orchestration:** Kubernetes - Container orchestration

## Casbin Model

The service uses RBAC model for permission control:

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

For detailed explanation, refer to [API.md#casbin-model](API.md#casbin-model).

## Integration with Controller

The Controller service can integrate with the RBAC service using the generated Golang client:

```go
import (
    rbac "github.com/davidliyutong/idekube-rbac/client"
)

// Create client
client, err := rbac.NewClient("http://rbac-service:8080")

// Check permission
allowed, err := client.CheckPermission(ctx, &rbac.CheckPermissionRequest{
    UserID:       userID,
    ResourceType: "workspace",
    ResourceID:   workspaceID,
    Action:       "delete",
})
```

For detailed integration guide, refer to [API.md#integration-guide](API.md#integration-guide).

## Monitoring and Observability

### Health Check

```bash
curl http://localhost:8080/healthz
```

### Logging

The service outputs structured logs with support for different log levels:

```bash
# Enable debug logging
export LOG_LEVEL=debug
```

### Metrics

It's recommended to integrate Prometheus for metric collection:

- HTTP request latency
- Permission check success/failure rate
- Database connection pool status
- RabbitMQ queue depth

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details

## Support

- 📧 Email: support@idekube.io
- 🐛 Issues: [GitHub Issues](https://github.com/davidliyutong/idekube/issues)
- 📖 Wiki: [Project Wiki](https://github.com/davidliyutong/idekube/wiki)

## Related Projects

- [idekube-controller](../controller/) - IDEKube controller service
- [idekube-housekeeper](../housekeeper/) - IDEKube cleanup service
- [idekube-frontend](../frontend/) - IDEKube frontend application
