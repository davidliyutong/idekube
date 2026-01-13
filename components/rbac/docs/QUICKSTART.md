# Quick Start Guide

This guide will help you quickly deploy and run the idekube-rbac service.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)
- [Verifying Deployment](#verifying-deployment)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Components

- **Go 1.21+** (for local development)
- **Docker** (for containerized deployment)
- **Kubernetes 1.24+** (for K8s deployment)
- **PostgreSQL 12+** (database)
- **RabbitMQ 3.9+** (message queue)

### Optional Components

- **Helm 3.0+** (for Kubernetes deployment)
- **Make** (build tool)
- **kubectl** (Kubernetes command-line tool)

## Local Development

### 1. Clone the Repository

```bash
git clone https://github.com/davidliyutong/idekube.git
cd idekube/components/rbac
```

### 2. Install Dependencies

```bash
make deps
```

### 3. Prepare Database

Start PostgreSQL (using Docker):

```bash
docker run -d \
  --name idekube-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=idekube_rbac \
  -p 5432:5432 \
  postgres:14-alpine
```

### 4. Prepare Message Queue

Start RabbitMQ (using Docker):

```bash
docker run -d \
  --name idekube-rabbitmq \
  -e RABBITMQ_DEFAULT_USER=guest \
  -e RABBITMQ_DEFAULT_PASS=guest \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3.12-management-alpine
```

### 5. Configure Environment Variables

Create a `.env` file:

```bash
# PostgreSQL configuration
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DB=idekube_rbac

# RabbitMQ configuration
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
export RABBITMQ_VHOST=/

# Application configuration
export HTTP_PORT=8080
export LOG_LEVEL=info
export WORKER_THREADS=1

# Casbin configuration
export CASBIN_MODEL_PATH=configs/model.conf
export CASBIN_POLICY_PATH=configs/policy.csv
```

Load environment variables:

```bash
source .env
```

### 6. Initialize Casbin Configuration

Create `configs/model.conf`:

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

Create `configs/policy.csv` (example policy):

```csv
p, role:admin, workspace, read
p, role:admin, workspace, write
p, role:admin, workspace, delete
p, role:admin, template, read
p, role:admin, template, write
p, role:admin, template, delete
p, role:editor, workspace, read
p, role:editor, workspace, write
p, role:viewer, workspace, read

g, user:1, role:admin
```

### 7. Build and Run

```bash
# Build
make build

# Run
make run
```

The service will start at `http://localhost:8080`.

### 8. Verify

```bash
# Health check
curl http://localhost:8080/healthz

# Access Swagger UI
open http://localhost:8080/swagger/
```

## Docker Deployment

### 1. Build Docker Image

```bash
make docker-build
```

This will create the `davidliyutong/idekube-rbac:latest` image.

### 2. Use Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: idekube_rbac
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  rbac:
    image: davidliyutong/idekube-rbac:latest
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    environment:
      POSTGRES_HOST: postgres
      POSTGRES_PORT: 5432
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: idekube_rbac
      RABBITMQ_HOST: rabbitmq
      RABBITMQ_PORT: 5672
      RABBITMQ_USER: guest
      RABBITMQ_PASSWORD: guest
      RABBITMQ_VHOST: /
      HTTP_PORT: 8080
      LOG_LEVEL: info
      CASBIN_MODEL_PATH: /app/configs/model.conf
      CASBIN_POLICY_PATH: /app/configs/policy.csv
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs

volumes:
  postgres_data:
  rabbitmq_data:
```

### 3. Start Services

```bash
docker-compose up -d
```

### 4. View Logs

```bash
docker-compose logs -f rbac
```

### 5. Stop Services

```bash
docker-compose down
```

## Kubernetes Deployment

### Using Helm Chart

IDEKube provides a complete Helm Chart for deployment.

#### 1. Prepare Values File

Create `values-rbac.yaml`:

```yaml
rbac:
  enabled: true
  image:
    repository: davidliyutong/idekube-rbac
    tag: latest
    pullPolicy: IfNotPresent
  
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  
  replicas: 2
  
  service:
    type: ClusterIP
    port: 8080
  
  env:
    LOG_LEVEL: info
    WORKER_THREADS: "1"
    HTTP_PORT: "8080"

postgresql:
  enabled: true
  auth:
    username: postgres
    password: password
    database: idekube_rbac
  primary:
    persistence:
      enabled: true
      size: 10Gi

rabbitmq:
  enabled: true
  auth:
    username: guest
    password: guest
  persistence:
    enabled: true
    size: 8Gi
```

#### 2. Install Chart

```bash
# From project root directory
cd /path/to/idekube

# Install
helm install idekube-rbac ./helm \
  -f values-rbac.yaml \
  -n idekube \
  --create-namespace
```

#### 3. Verify Deployment

```bash
# Check Pod status
kubectl get pods -n idekube

# Check services
kubectl get svc -n idekube

# View logs
kubectl logs -n idekube -l app=idekube-rbac -f
```

#### 4. Access Service

```bash
# Port forward
kubectl port-forward -n idekube svc/idekube-rbac 8080:8080

# Access
curl http://localhost:8080/healthz
```

### Using Manifest Files

If not using Helm, you can use raw Kubernetes manifests:

#### 1. Create Namespace

```bash
kubectl create namespace idekube
```

#### 2. Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rbac-config
  namespace: idekube
data:
  model.conf: |
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
  
  policy.csv: |
    p, role:admin, workspace, read
    p, role:admin, workspace, write
    p, role:admin, workspace, delete
```

```bash
kubectl apply -f rbac-configmap.yaml
```

#### 3. Create Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rbac-secrets
  namespace: idekube
type: Opaque
stringData:
  POSTGRES_PASSWORD: password
  RABBITMQ_PASSWORD: guest
```

```bash
kubectl apply -f rbac-secret.yaml
```

#### 4. Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idekube-rbac
  namespace: idekube
spec:
  replicas: 2
  selector:
    matchLabels:
      app: idekube-rbac
  template:
    metadata:
      labels:
        app: idekube-rbac
    spec:
      containers:
      - name: rbac
        image: davidliyutong/idekube-rbac:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: POSTGRES_HOST
          value: postgresql
        - name: POSTGRES_PORT
          value: "5432"
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rbac-secrets
              key: POSTGRES_PASSWORD
        - name: POSTGRES_DB
          value: idekube_rbac
        - name: RABBITMQ_HOST
          value: rabbitmq
        - name: RABBITMQ_PORT
          value: "5672"
        - name: RABBITMQ_USER
          value: guest
        - name: RABBITMQ_PASSWORD
          valueFrom:
            secretKeyRef:
              name: rbac-secrets
              key: RABBITMQ_PASSWORD
        - name: HTTP_PORT
          value: "8080"
        - name: LOG_LEVEL
          value: info
        - name: CASBIN_MODEL_PATH
          value: /etc/rbac/model.conf
        - name: CASBIN_POLICY_PATH
          value: /etc/rbac/policy.csv
        volumeMounts:
        - name: config
          mountPath: /etc/rbac
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: rbac-config
```

```bash
kubectl apply -f rbac-deployment.yaml
```

#### 5. Create Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: idekube-rbac
  namespace: idekube
spec:
  selector:
    app: idekube-rbac
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
    name: http
  type: ClusterIP
```

```bash
kubectl apply -f rbac-service.yaml
```

## Configuration

### Environment Variables

| Variable Name | Required | Default | Description |
|--------------|----------|---------|-------------|
| POSTGRES_HOST | Yes | localhost | PostgreSQL host address |
| POSTGRES_PORT | Yes | 5432 | PostgreSQL port |
| POSTGRES_USER | Yes | - | PostgreSQL username |
| POSTGRES_PASSWORD | Yes | - | PostgreSQL password |
| POSTGRES_DB | Yes | - | PostgreSQL database name |
| RABBITMQ_HOST | Yes | localhost | RabbitMQ host address |
| RABBITMQ_PORT | Yes | 5672 | RabbitMQ port |
| RABBITMQ_USER | Yes | - | RabbitMQ username |
| RABBITMQ_PASSWORD | Yes | - | RabbitMQ password |
| RABBITMQ_VHOST | No | / | RabbitMQ virtual host |
| HTTP_PORT | No | 8080 | HTTP service port |
| LOG_LEVEL | No | info | Log level (debug/info/warn/error) |
| WORKER_THREADS | No | 1 | Number of worker threads |
| CASBIN_MODEL_PATH | Yes | - | Casbin model configuration file path |
| CASBIN_POLICY_PATH | No | - | Casbin policy file path (optional) |

### Casbin Configuration

#### Model File (model.conf)

Defines the structure of RBAC rules. The default configuration provided above is recommended.

#### Policy File (policy.csv)

Defines specific permission policies. Format:

```csv
# Policy rules: p, subject, object, action
p, role:admin, workspace, read
p, role:admin, workspace, write

# Role inheritance: g, user, role
g, user:1, role:admin
g, user:2, role:editor
```

**Note:** Policies can also be stored in the PostgreSQL database, allowing dynamic management without service restart.

## Verifying Deployment

### 1. Health Check

```bash
curl http://localhost:8080/healthz
```

Expected output: `ok`

### 2. Test Permission Check

```bash
curl -X POST http://localhost:8080/api/v1/rbac/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "resource_type": "workspace",
    "action": "read"
  }'
```

Expected output: `{"allowed":true}` or `{"allowed":false}`

### 3. Access Swagger UI

Open in browser: `http://localhost:8080/swagger/`

### 4. Check Logs

#### Docker
```bash
docker logs idekube-rbac
```

#### Kubernetes
```bash
kubectl logs -n idekube -l app=idekube-rbac
```

## Troubleshooting

### Service Won't Start

**Symptom:** Service fails to start or exits immediately

**Checklist:**

1. Verify PostgreSQL connection
   ```bash
   psql -h localhost -U postgres -d idekube_rbac
   ```

2. Verify RabbitMQ connection
   ```bash
   curl http://localhost:15672/api/overview
   ```

3. Check if environment variables are correctly set

4. View detailed logs (set `LOG_LEVEL=debug`)

### Database Connection Failed

**Error:** `Failed to connect to PostgreSQL`

**Solution:**

1. Ensure PostgreSQL is running
   ```bash
   docker ps | grep postgres
   ```

2. Check connection string
   ```bash
   psql "postgresql://postgres:password@localhost:5432/idekube_rbac"
   ```

3. Verify network connectivity
   ```bash
   telnet localhost 5432
   ```

### RabbitMQ Connection Failed

**Error:** `Failed to connect to RabbitMQ`

**Solution:**

1. Ensure RabbitMQ is running
   ```bash
   docker ps | grep rabbitmq
   ```

2. Check management interface
   ```bash
   curl http://localhost:15672
   ```

3. Verify credentials
   ```bash
   rabbitmqctl authenticate_user guest guest
   ```

### Casbin Initialization Failed

**Error:** `Failed to initialize RBAC service`

**Solution:**

1. Check if model file exists
   ```bash
   ls -la configs/model.conf
   ```

2. Verify model file syntax

3. Check file permissions

### Kubernetes Pod Stuck in CrashLoopBackOff

**Solution:**

1. View Pod details
   ```bash
   kubectl describe pod -n idekube <pod-name>
   ```

2. View logs
   ```bash
   kubectl logs -n idekube <pod-name> --previous
   ```

3. Check if ConfigMap and Secret are correctly mounted

4. Verify dependent services (PostgreSQL, RabbitMQ) are ready

### Slow API Response

**Possible Causes:**

1. Insufficient database connection pool
2. Database missing indexes
3. Too many policy rules

**Optimization Suggestions:**

1. Increase database connection pool size
2. Add indexes to Casbin policy table
3. Enable result caching
4. Increase replica count (Kubernetes)

## Production Environment Recommendations

### 1. Security

- Use strong passwords
- Enable TLS/HTTPS
- Configure network policies
- Regularly update dependencies

### 2. High Availability

- Deploy multiple replicas (at least 2)
- Configure Pod anti-affinity
- Use HPA for auto-scaling
- Configure health checks

### 3. Monitoring

- Configure Prometheus monitoring
- Set up alert rules
- Collect logs to centralized logging system
- Monitor key metrics (latency, error rate, throughput)

### 4. Backup

- Regularly backup PostgreSQL database
- Backup Casbin configuration files
- Test recovery procedures

## Next Steps

- Read [API Definition Documentation](API.md) for API details
- Check out [API Testing Guide](API_TESTING.md) to learn how to test
- Browse [README](README.md) for project overview

## Getting Help

- Check project Issues
- Read project Wiki
- Contact development team

---

**Enjoy using the service!**
