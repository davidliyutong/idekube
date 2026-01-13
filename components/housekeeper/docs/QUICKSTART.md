# Quick Start Guide

This document provides deployment and configuration guidelines for the idekube-housekeeper component.

## Overview

idekube-housekeeper is the cleanup and maintenance component of the idekube platform, responsible for:

- Cleaning up expired Kubernetes workspace resources
- Archiving workspace data
- Executing periodic maintenance tasks
- Receiving cleanup requests via RabbitMQ

## System Requirements

### Runtime Environment

- Kubernetes 1.20+
- PostgreSQL 12+
- RabbitMQ 3.8+
- Go 1.21+ (development only)

### Resource Quotas

**Minimum Configuration**:
- CPU: 100m
- Memory: 128Mi
- Storage: No special requirements

**Recommended Configuration**:
- CPU: 500m
- Memory: 512Mi
- Storage: Based on logging needs

**Production Environment**:
- CPU: 1000m (1 core)
- Memory: 1Gi
- Storage: 10Gi (for logs and temporary files)

## Local Development

### 1. Clone the Repository

```bash
git clone https://github.com/davidliyutong/idekube.git
cd idekube/components/housekeeper
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Create a `.env` file:

```bash
# Kubernetes Configuration
export KUBECONFIG=~/.kube/config
export NAMESPACE=default

# PostgreSQL Configuration
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=idekube
export POSTGRES_PASSWORD=your_password
export POSTGRES_DB=idekube

# RabbitMQ Configuration
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
export RABBITMQ_VHOST=/

# Application Configuration
export LOG_LEVEL=debug
export WORKER_THREADS=1
```

Load environment variables:

```bash
source .env
```

### 4. Start Dependent Services

#### Using Docker Compose

Create `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: idekube-postgres
    environment:
      POSTGRES_USER: idekube
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: idekube
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U idekube"]
      interval: 10s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: idekube-rabbitmq
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

volumes:
  postgres_data:
  rabbitmq_data:
```

Start services:

```bash
docker-compose -f docker-compose.dev.yml up -d
```

Verify service status:

```bash
# Check PostgreSQL
docker exec idekube-postgres pg_isready -U idekube

# Check RabbitMQ
docker exec idekube-rabbitmq rabbitmq-diagnostics ping
```

### 5. Initialize Database

Run database migrations (if migration tool exists):

```bash
# Example: using migrate tool
migrate -path ./migrations -database "postgresql://idekube:your_password@localhost:5432/idekube?sslmode=disable" up
```

Or manually execute SQL:

```bash
psql -h localhost -U idekube -d idekube -f ./migrations/000001_init_schema.up.sql
```

### 6. Build and Run

```bash
# Build
make build

# Run
make run
```

Or run directly with go:

```bash
go run cmd/housekeeper/main.go
```

### 7. Verify Running Status

Check log output:

```
INFO  Starting idekube-housekeeper
INFO  Configuration loaded successfully
INFO  Connected to PostgreSQL at localhost:5432
INFO  Connected to RabbitMQ at localhost:5672
INFO  Kubernetes client initialized
INFO  Housekeeper started
DEBUG Housekeeper heartbeat
```

## Docker Deployment

### 1. Build Image

```bash
# Using Makefile
make docker-build

# Or directly using docker
docker build -t idekube/housekeeper:latest .
```

### 2. Push Image

```bash
# Tag image
docker tag idekube/housekeeper:latest your-registry.com/idekube/housekeeper:latest

# Push to image registry
docker push your-registry.com/idekube/housekeeper:latest
```

### 3. Run Container

```bash
docker run -d \
  --name idekube-housekeeper \
  --network host \
  -e POSTGRES_HOST=localhost \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=idekube \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=idekube \
  -e RABBITMQ_HOST=localhost \
  -e RABBITMQ_PORT=5672 \
  -e RABBITMQ_USER=guest \
  -e RABBITMQ_PASSWORD=guest \
  -e LOG_LEVEL=info \
  -v ~/.kube/config:/root/.kube/config:ro \
  idekube/housekeeper:latest
```

### 4. View Logs

```bash
docker logs -f idekube-housekeeper
```

### 5. Stop and Cleanup

```bash
# Stop container
docker stop idekube-housekeeper

# Remove container
docker rm idekube-housekeeper
```

## Kubernetes Deployment

### Method 1: Using kubectl

#### 1. Create Namespace

```bash
kubectl create namespace idekube-system
```

#### 2. Create Secret

```bash
kubectl create secret generic housekeeper-secrets \
  --from-literal=postgres-user=idekube \
  --from-literal=postgres-password=your_password \
  --from-literal=rabbitmq-user=guest \
  --from-literal=rabbitmq-password=guest \
  -n idekube-system
```

#### 3. Create Deployment

Create `deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idekube-housekeeper
  namespace: idekube-system
  labels:
    app: idekube-housekeeper
spec:
  replicas: 1
  selector:
    matchLabels:
      app: idekube-housekeeper
  template:
    metadata:
      labels:
        app: idekube-housekeeper
    spec:
      serviceAccountName: idekube-housekeeper
      containers:
      - name: housekeeper
        image: idekube/housekeeper:latest
        imagePullPolicy: Always
        env:
        - name: NAMESPACE
          value: "idekube-system"
        - name: POSTGRES_HOST
          value: "postgresql"
        - name: POSTGRES_PORT
          value: "5432"
        - name: POSTGRES_DB
          value: "idekube"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: postgres-password
        - name: RABBITMQ_HOST
          value: "rabbitmq"
        - name: RABBITMQ_PORT
          value: "5672"
        - name: RABBITMQ_VHOST
          value: "/"
        - name: RABBITMQ_USER
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: rabbitmq-user
        - name: RABBITMQ_PASSWORD
          valueFrom:
            secretKeyRef:
              name: housekeeper-secrets
              key: rabbitmq-password
        - name: LOG_LEVEL
          value: "info"
        - name: WORKER_THREADS
          value: "1"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pgrep -f idekube-housekeeper
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pgrep -f idekube-housekeeper
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### 4. Create ServiceAccount and RBAC

Create `rbac.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: idekube-housekeeper
  namespace: idekube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: idekube-housekeeper
rules:
- apiGroups: [""]
  resources: ["namespaces", "pods", "services", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "replicasets"]
  verbs: ["get", "list", "watch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: idekube-housekeeper
subjects:
- kind: ServiceAccount
  name: idekube-housekeeper
  namespace: idekube-system
roleRef:
  kind: ClusterRole
  name: idekube-housekeeper
  apiGroup: rbac.authorization.k8s.io
```

#### 5. Apply Configuration

```bash
# Apply RBAC
kubectl apply -f rbac.yaml

# Apply Deployment
kubectl apply -f deployment.yaml
```

#### 6. Verify Deployment

```bash
# Check Pod status
kubectl get pods -n idekube-system -l app=idekube-housekeeper

# View logs
kubectl logs -f deployment/idekube-housekeeper -n idekube-system

# Check ServiceAccount
kubectl get serviceaccount idekube-housekeeper -n idekube-system
```

### Method 2: Using Helm Chart

#### 1. Prepare values.yaml

Create `values.yaml` (or use the Helm chart in project root):

```yaml
housekeeper:
  enabled: true
  replicas: 1
  
  image:
    repository: idekube/housekeeper
    tag: latest
    pullPolicy: Always
  
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  
  env:
    logLevel: info
    workerThreads: 1
  
  postgres:
    host: postgresql
    port: 5432
    database: idekube
    existingSecret: housekeeper-secrets
    userKey: postgres-user
    passwordKey: postgres-password
  
  rabbitmq:
    host: rabbitmq
    port: 5672
    vhost: /
    existingSecret: housekeeper-secrets
    userKey: rabbitmq-user
    passwordKey: rabbitmq-password
  
  serviceAccount:
    create: true
    name: idekube-housekeeper
```

#### 2. Install Helm Chart

```bash
# From project root
cd /path/to/idekube

# Install or upgrade
helm upgrade --install idekube ./helm \
  --namespace idekube-system \
  --create-namespace \
  -f values.yaml
```

#### 3. Verify Installation

```bash
# Check release status
helm status idekube -n idekube-system

# View all resources
kubectl get all -n idekube-system -l app=idekube-housekeeper
```

#### 4. Uninstall

```bash
helm uninstall idekube -n idekube-system
```

## Configuration

### Environment Variables

#### Kubernetes Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| KUBECONFIG | - | Path to kubeconfig file (can be omitted when running in-cluster) |
| NAMESPACE | Empty (all namespaces) | Namespace to watch |

#### PostgreSQL Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| POSTGRES_HOST | localhost | PostgreSQL host address |
| POSTGRES_PORT | 5432 | PostgreSQL port |
| POSTGRES_USER | - | Database username (required) |
| POSTGRES_PASSWORD | - | Database password (required) |
| POSTGRES_DB | - | Database name (required) |

Connection string format:
```
postgresql://<user>:<password>@<host>:<port>/<database>?sslmode=disable
```

#### RabbitMQ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| RABBITMQ_HOST | localhost | RabbitMQ host address |
| RABBITMQ_PORT | 5672 | RabbitMQ port |
| RABBITMQ_USER | - | RabbitMQ username (required) |
| RABBITMQ_PASSWORD | - | RabbitMQ password (required) |
| RABBITMQ_VHOST | / | RabbitMQ virtual host |

#### Application Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| LOG_LEVEL | info | Log level: debug, info, warn, error |
| WORKER_THREADS | 1 | Number of worker threads |

### Log Levels

- **debug**: Detailed debug information including all operation details
- **info**: Regular operational information including startup, shutdown, and important events
- **warn**: Warning messages indicating potential issues that don't affect operation
- **error**: Error messages indicating failed operations requiring attention

Recommended settings:
- Development: `debug`
- Testing: `info`
- Production: `info` or `warn`

## Health Checks

### Liveness Probe

Check if the process is alive:

```yaml
livenessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - pgrep -f idekube-housekeeper
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe

Check if the service is ready:

```yaml
readinessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - pgrep -f idekube-housekeeper
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## Monitoring and Logging

### Log Collection

#### Using kubectl

```bash
# View logs in real-time
kubectl logs -f deployment/idekube-housekeeper -n idekube-system

# View last 100 lines of logs
kubectl logs --tail=100 deployment/idekube-housekeeper -n idekube-system

# View logs of a specific Pod
kubectl logs <pod-name> -n idekube-system
```

#### Integration with Logging Systems

**Fluentd/Fluent Bit**:
```yaml
# Example: Add log labels
metadata:
  annotations:
    fluentd.io/parse: "json"
    fluentd.io/exclude: "false"
```

**ELK Stack**:
- Configure Filebeat to collect container logs
- Use JSON format for easier parsing

### Monitoring Metrics

Recommended monitoring metrics:

1. **Application Metrics**:
   - CPU usage
   - Memory usage
   - Goroutine count
   - Message processing rate

2. **Business Metrics**:
   - Cleanup task success/failure count
   - Average cleanup time
   - Queue backlog count
   - Database connection count

3. **Dependency Services**:
   - PostgreSQL connection status
   - RabbitMQ connection status
   - Kubernetes API response time

## Troubleshooting

### Common Issues

#### 1. Cannot Connect to PostgreSQL

**Error message**:
```
FATAL Failed to connect to PostgreSQL: connection refused
```

**Troubleshooting steps**:
```bash
# Check PostgreSQL service
kubectl get svc postgresql -n idekube-system

# Test connection
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql -h postgresql -U idekube -d idekube -c "SELECT 1;"

# Check network policies
kubectl get networkpolicies -n idekube-system
```

#### 2. Cannot Connect to RabbitMQ

**Error message**:
```
FATAL Failed to connect to RabbitMQ: connection refused
```

**Troubleshooting steps**:
```bash
# Check RabbitMQ service
kubectl get svc rabbitmq -n idekube-system

# Test connection
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl -u guest:guest http://rabbitmq:15672/api/overview

# Check RabbitMQ logs
kubectl logs deployment/rabbitmq -n idekube-system
```

#### 3. Insufficient Permissions

**Error message**:
```
ERROR Failed to delete namespace: forbidden
```

**Troubleshooting steps**:
```bash
# Check ServiceAccount
kubectl get serviceaccount idekube-housekeeper -n idekube-system

# Check ClusterRoleBinding
kubectl get clusterrolebinding idekube-housekeeper

# Verify permissions
kubectl auth can-i delete namespaces \
  --as=system:serviceaccount:idekube-system:idekube-housekeeper
```

#### 4. Pod Startup Failure

```bash
# View Pod status
kubectl describe pod -l app=idekube-housekeeper -n idekube-system

# View events
kubectl get events -n idekube-system --sort-by='.lastTimestamp'

# Check image pull
kubectl get pods -n idekube-system -o jsonpath='{.items[*].status.containerStatuses[*].state}'
```

### Debug Mode

Enable debug logging:

```bash
# Temporarily modify environment variable
kubectl set env deployment/idekube-housekeeper LOG_LEVEL=debug -n idekube-system

# View detailed logs
kubectl logs -f deployment/idekube-housekeeper -n idekube-system
```

## Upgrade and Rollback

### Upgrade

```bash
# Use new image
kubectl set image deployment/idekube-housekeeper \
  housekeeper=idekube/housekeeper:v1.1.0 \
  -n idekube-system

# Check upgrade status
kubectl rollout status deployment/idekube-housekeeper -n idekube-system
```

### Rollback

```bash
# View version history
kubectl rollout history deployment/idekube-housekeeper -n idekube-system

# Rollback to previous version
kubectl rollout undo deployment/idekube-housekeeper -n idekube-system

# Rollback to specific version
kubectl rollout undo deployment/idekube-housekeeper --to-revision=2 -n idekube-system
```

## Security Recommendations

1. **Use Secrets for Sensitive Information**:
   - Don't hardcode passwords in configuration files
   - Use Kubernetes Secrets or external secret management systems

2. **Principle of Least Privilege**:
   - Grant ServiceAccount only necessary permissions
   - Use RBAC to limit resource access scope

3. **Network Isolation**:
   - Use NetworkPolicy to restrict network access
   - Allow only necessary inter-service communication

4. **Image Security**:
   - Use official base images
   - Regularly update dependencies and base images
   - Scan images for vulnerabilities

5. **Log Security**:
   - Don't output sensitive information in logs
   - Set appropriate log retention policies

## Performance Optimization

1. **Resource Quotas**:
   - Adjust CPU and memory limits based on actual workload
   - Use HPA for horizontal scaling (if needed)

2. **Concurrent Processing**:
   - Increase WORKER_THREADS to improve concurrency
   - Pay attention to database connection pool size

3. **Batch Operations**:
   - Batch delete Kubernetes resources
   - Batch update database records

4. **Caching Strategy**:
   - Cache frequently used Kubernetes resource information
   - Reduce database query frequency

## Next Steps

- Read [API Definition](./API.md) to understand RabbitMQ message formats
- Read [API Testing Guide](./API_TESTING.md) to learn how to test functionality
- Check [Main README](./README.md) for project architecture

## Getting Help

- GitHub Issues: https://github.com/davidliyutong/idekube/issues
- Project Documentation: https://github.com/davidliyutong/idekube
