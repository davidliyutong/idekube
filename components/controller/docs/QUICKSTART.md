````markdown
# IDEKube Controller Quick Start Guide

This guide helps you quickly deploy and run the IDEKube Controller service.

> 📖 **Related Documentation**
> - [API Documentation](./API.md) - Complete API specifications
> - [API Testing Guide](./API_TESTING.md) - API testing examples
> - [Swagger UI](http://localhost:8080/swagger/index.html) - Interactive API documentation

## System Requirements

### Required Components

- **Go**: 1.21 or higher
- **PostgreSQL**: 14 or higher
- **Kubernetes**: 1.24+ (optional, for full workspace functionality)

### Optional Components

- **RabbitMQ**: 3.12+ (for asynchronous event processing)
- **Docker**: For containerized deployment

## Option 1: Local Development

### 1. Install PostgreSQL

#### macOS
```bash
brew install postgresql@14
brew services start postgresql@14
```

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install postgresql-14
sudo systemctl start postgresql
```

### 2. Create Database

```bash
# Switch to postgres user
sudo -u postgres psql

# Execute in psql
CREATE DATABASE idekube;
CREATE USER idekube WITH PASSWORD 'idekube123';
GRANT ALL PRIVILEGES ON DATABASE idekube TO idekube;
\q
```

Or use command line:
```bash
createdb idekube
createuser -P idekube  # Enter password: idekube123
psql -c "GRANT ALL PRIVILEGES ON DATABASE idekube TO idekube;"
```

> ⚠️ **Note**: Use strong passwords in production!

### 3. Configure Environment Variables

```bash
cd components/controller
cp .env.example .env
```

Edit the `.env` file:

```bash
# Server configuration
SERVER_ADDRESS=:8080

# JWT configuration
JWT_SECRET=your-super-secret-key-change-this-in-production
JWT_EXPIRATION_HOURS=720

# PostgreSQL configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=idekube
POSTGRES_PASSWORD=idekube123
POSTGRES_DB=idekube

# Kubernetes configuration (optional)
KUBECONFIG=~/.kube/config
NAMESPACE=idekube

# RabbitMQ configuration (optional)
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run Service

```bash
# Using make (recommended)
make run

# Or use go run directly
go run cmd/controller/main.go
```

After service starts, it will:
1. Connect to PostgreSQL database
2. **Automatically run database migrations** (no need to manually execute SQL)
3. Start HTTP server: http://localhost:8080

### 6. Verify Service

**Health check:**
```bash
curl http://localhost:8080/health
```

**View Swagger documentation:**

Open in browser: http://localhost:8080/swagger/index.html

**Register test user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "full_name": "Admin User"
  }'
```

## Option 2: Docker Deployment

### 1. Build Image

```bash
make docker-build
```

### 2. Run with Docker

**Run controller only (requires external PostgreSQL):**
```bash
docker run --rm \
  -e POSTGRES_HOST=host.docker.internal \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=idekube \
  -e POSTGRES_PASSWORD=idekube123 \
  -e POSTGRES_DB=idekube \
  -e JWT_SECRET=your-secret-key \
  -p 8080:8080 \
  davidliyutong/idekube-controller:latest
```

**Using Docker Compose (recommended):**

Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: idekube
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: idekube123
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U idekube"]
      interval: 5s
      timeout: 5s
      retries: 5

  controller:
    image: davidliyutong/idekube-controller:latest
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      SERVER_ADDRESS: ":8080"
      JWT_SECRET: "your-secret-key-change-in-production"
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: idekube123
      POSTGRES_DB: idekube
    ports:
      - "8080:8080"

volumes:
  postgres-data:
```

Start services:
```bash
docker-compose up -d
```

## Option 3: Kubernetes Deployment

### Using Helm Chart

```bash
helm install idekube-controller ../../helm \
  --set controller.image.tag=latest \
  --set postgresql.enabled=true
```

### Using Kubectl

```bash
kubectl apply -f manifests/
kubectl get pods -n idekube
```

## Generate Swagger Documentation

```bash
# Generate Swagger documentation
make swagger-gen

# Generate JavaScript client
make swagger-js-client

# Generate TypeScript client
make swagger-ts-client
```

## Development Tools

### Hot Reload Development

```bash
go install github.com/cosmtrek/air@latest
make dev
```

### Code Quality

```bash
make fmt          # Format code
make lint         # Lint code
make test         # Run tests
make test-coverage # Test coverage
```

## Common Issues

### PostgreSQL Connection Failed

Check if PostgreSQL is running: `pg_isready`

Docker users should use `host.docker.internal` to connect to host machine.

### Port Already in Use

```bash
lsof -i :8080  # Find process using port
export SERVER_ADDRESS=:8081  # Or change port
```

### Swagger Documentation Not Generated

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger-gen
```

## Next Steps

- 📖 [API Documentation](./API.md) - Complete API specifications
- 🧪 [API Testing Guide](./API_TESTING.md) - Testing examples
- 🔍 [Swagger UI](http://localhost:8080/swagger/index.html) - Interactive documentation

### Phase 1 Implementation (TODO)

The following features are designed but need handlers and services to be implemented:

1. **Organization Management** - Create and manage organizations
2. **Template Management** - Define workspace templates
3. **Workspace Management** - Create and orchestrate workspaces in K8s
4. **Volume Management** - Manage persistent volumes

Refer to [DESIGN.md](DESIGN.md) for detailed API specifications and implementation guidance.

### Development Workflow

1. Implement service layer in `internal/services/`
2. Implement handler layer in `internal/handlers/`
3. Register routes in `internal/api/server.go`
4. Add K8s orchestration logic in `pkg/k8s/`
5. Test the endpoints

## Troubleshooting

### Database connection failed
- Check PostgreSQL is running: `pg_isready`
- Verify credentials in `.env`
- Check database exists: `psql -l`

### JWT token invalid
- Check `JWT_SECRET` is set correctly
- Verify token hasn't expired (default 24 hours)
- Ensure Bearer token format: `Authorization: Bearer <token>`

### Permission denied
- Check user role (admin/super_admin required for some endpoints)
- Verify JWT token is valid and not expired

## Docker Deployment

```bash
# Build image
make docker-build

# Run container
docker run -d \
  --name idekube-controller \
  -p 8080:8080 \
  --env-file .env \
  idekube/controller:latest
```

## Production Deployment Checklist

- [ ] Change default admin password
- [ ] Set strong JWT_SECRET (32+ random characters)
- [ ] Use environment-specific database credentials
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring and alerts
- [ ] Review and adjust quotas
- [ ] Enable audit log retention policies

## Support

For issues and questions:
- Check [DESIGN.md](DESIGN.md) for architecture details
- Review API documentation in code comments
- Check logs for error messages

````