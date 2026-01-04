# idekube-rbac

Kubernetes RBAC service for managing access control in the idekube platform.

## Features

- Kubernetes CRD operations
- PostgreSQL integration
- RabbitMQ integration
- Cloud-native configuration via environment variables

## Configuration

The controller is configured via environment variables:

### Kubernetes
- `KUBECONFIG`: Path to kubeconfig file (optional when running in-cluster)
- `NAMESPACE`: Namespace to watch (default: all namespaces)

### PostgreSQL
- `POSTGRES_HOST`: PostgreSQL host (default: localhost)
- `POSTGRES_PORT`: PostgreSQL port (default: 5432)
- `POSTGRES_USER`: PostgreSQL user
- `POSTGRES_PASSWORD`: PostgreSQL password
- `POSTGRES_DB`: PostgreSQL database name

### RabbitMQ
- `RABBITMQ_HOST`: RabbitMQ host (default: localhost)
- `RABBITMQ_PORT`: RabbitMQ port (default: 5672)
- `RABBITMQ_USER`: RabbitMQ user
- `RABBITMQ_PASSWORD`: RabbitMQ password
- `RABBITMQ_VHOST`: RabbitMQ virtual host (default: /)

### Application
- `LOG_LEVEL`: Log level (default: info)
- `WORKER_THREADS`: Number of worker threads (default: 1)

## Development

### Prerequisites
- Go 1.21+
- Docker (for building images)
- Kubernetes cluster (for testing)

### Build
```bash
make build
```

### Run locally
```bash
make run
```

### Build Docker image
```bash
make docker-build
```

### Run tests
```bash
make test
```

## Deployment

Deploy using the provided Helm chart in the parent directory.
