# idekube Helm Chart

A Kubernetes-native IDE platform with built-in RBAC, container lifecycle management, and web-based interface.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start (Testing)](#quick-start-testing)
- [Production Deployment](#production-deployment)
- [Configuration](#configuration)
- [Backup and Recovery](#backup-and-recovery)
- [Contributing](#contributing)
- [License](#license)

## Prerequisites

- Kubernetes cluster (v1.23+)
- Helm 3.x
- kubectl configured and connected to your cluster
- OpenSSL (for generating secure credentials)

## Quick Start (Testing)

⚠️ **Note**: This is for testing/development only. For production, see [Production Deployment](#production-deployment).

### 1. Validate Chart

```bash
# Lint the chart
helm lint .

# Preview rendered templates
helm template idekube . --debug
```

### 2. Generate Test Credentials

Even for testing, generate secure random credentials instead of using hardcoded values:

```bash
# Generate test credentials
export TEST_POSTGRES_PASS=$(openssl rand -base64 32)
export TEST_RABBITMQ_PASS=$(openssl rand -base64 32)
export TEST_ERLANG_COOKIE=$(openssl rand -base64 32)
export TEST_JWT_SECRET=$(openssl rand -base64 64)
```

### 3. Install to Test Namespace

```bash
# Create test namespace
kubectl create namespace idekube-test

# Install with generated credentials
helm install idekube . \
  --namespace idekube-test \
  --set postgresql.auth.password="$TEST_POSTGRES_PASS" \
  --set rabbitmq.auth.password="$TEST_RABBITMQ_PASS" \
  --set rabbitmq.auth.erlangCookie="$TEST_ERLANG_COOKIE" \
  --set secrets.DATABASE_PASSWORD="$TEST_POSTGRES_PASS" \
  --set secrets.RABBITMQ_PASSWORD="$TEST_RABBITMQ_PASS" \
  --set secrets.JWT_SECRET="$TEST_JWT_SECRET" \
  --set ingress.enabled=false
```

### 4. Verify Deployment

```bash
# Check release status
helm status idekube -n idekube-test

# Watch pods come up
kubectl get pods -n idekube-test -w

# Check all resources
kubectl get all -n idekube-test
```

### 5. Access Services

```bash
# PostgreSQL (for debugging)
kubectl port-forward -n idekube-test svc/idekube-postgresql 5432:5432
# Connect: psql -h localhost -U idekube -d idekube

# RabbitMQ Management UI
kubectl port-forward -n idekube-test svc/idekube-rabbitmq 15672:15672
# Access: http://localhost:15672
# Username: idekube, Password: $TEST_RABBITMQ_PASS

# Frontend
kubectl port-forward -n idekube-test svc/idekube-frontend 8080:80
# Access: http://localhost:8080
```

### 6. View Logs

```bash
# Controller
kubectl logs -n idekube-test -l app.kubernetes.io/component=controller --tail=50

# RBAC service
kubectl logs -n idekube-test -l app.kubernetes.io/component=rbac --tail=50

# Housekeeper
kubectl logs -n idekube-test -l app.kubernetes.io/component=housekeeping --tail=50

# Frontend
kubectl logs -n idekube-test -l app.kubernetes.io/component=frontend --tail=50
```

### 7. Cleanup

```bash
# Uninstall release
helm uninstall idekube -n idekube-test

# Delete namespace and all resources
kubectl delete namespace idekube-test
```

## Production Deployment

⚠️ **CRITICAL SECURITY REQUIREMENTS**:

- All passwords and secrets MUST be unique and randomly generated
- Never use default or example values in production
- Store credentials securely (password manager, vault, etc.)
- Enable TLS/SSL for all external connections
- Use Kubernetes Secrets or external secret management

### Step 1: Generate Production Credentials

```bash
# Generate all production credentials
export POSTGRES_PASSWORD=$(openssl rand -base64 32)
export RABBITMQ_PASSWORD=$(openssl rand -base64 32)
export RABBITMQ_ERLANG_COOKIE=$(openssl rand -base64 32)
export JWT_SECRET=$(openssl rand -base64 64)

# Display generated values (save these securely!)
echo "PostgreSQL Password: $POSTGRES_PASSWORD"
echo "RabbitMQ Password: $RABBITMQ_PASSWORD"
echo "RabbitMQ Erlang Cookie: $RABBITMQ_ERLANG_COOKIE"
echo "JWT Secret: $JWT_SECRET"
```

### Step 2: Create Kubernetes Secrets

```bash
# Create production namespace
kubectl create namespace idekube-prod

# Create PostgreSQL secret
kubectl create secret generic postgresql-secret \
  --from-literal=postgres-password="$POSTGRES_PASSWORD" \
  --namespace idekube-prod

# Create RabbitMQ secret
kubectl create secret generic rabbitmq-secret \
  --from-literal=rabbitmq-password="$RABBITMQ_PASSWORD" \
  --from-literal=rabbitmq-erlang-cookie="$RABBITMQ_ERLANG_COOKIE" \
  --namespace idekube-prod
```

### Step 3: Create Production Values File

Create `values-prod.yaml` (without secrets):

```yaml
# Global configuration
global:
  storageClass: "your-storage-class"  # e.g., gp3, fast-ssd

# Controller - Core service
controller:
  image:
    tag: "v1.0.0"  # Use specific version, not "latest"

# RBAC service
rbac:
  image:
    tag: "v1.0.0"

# Frontend
frontend:
  image:
    tag: "v1.0.0"

# Ingress configuration
ingress:
  enabled: true
  className: "nginx"
  hostname: "idekube.yourdomain.com"  # CHANGE THIS
  certIssuer: "letsencrypt-prod"
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "100m"

# PostgreSQL
postgresql:
  auth:
    existingSecret: "postgresql-secret"
  primary:
    persistence:
      enabled: true
      size: 10Gi

# RabbitMQ
rabbitmq:
  auth:
    existingPasswordSecret: "rabbitmq-secret"
  persistence:
    enabled: true
    size: 5Gi
```

### Step 4: Deploy with Helm

```bash
# Install to production
helm install idekube . \
  --namespace idekube-prod \
  --values values-prod.yaml \
  --set secrets.JWT_SECRET="$JWT_SECRET" \
  --wait \
  --timeout 10m
```

#### Step 5: Verify Production Deployment

```bash
# Check all pods are running
kubectl get pods -n idekube-prod

# Check ingress
kubectl get ingress -n idekube-prod

# Verify services
kubectl get svc -n idekube-prod

# Check for any issues
kubectl get events -n idekube-prod --sort-by='.lastTimestamp'
```

### Post-Deployment Security Checklist

After deploying to production, verify:

- [ ] All passwords are unique and randomly generated (min 32 chars)
- [ ] No default values from `values.yaml` are being used
- [ ] Credentials are stored securely (not in git, not in plain text)
- [ ] TLS/SSL is enabled on ingress
- [ ] Database backups are configured
- [ ] Monitoring and logging are enabled
- [ ] Resource limits are set appropriately
- [ ] Network policies are in place (if required)
- [ ] RBAC permissions are configured correctly

### Upgrading Production Deployment

```bash
# Update with new values
helm upgrade idekube . \
  --namespace idekube-prod \
  --values values-prod.yaml \
  --reuse-values

# Or specify new image versions
helm upgrade idekube . \
  --namespace idekube-prod \
  --values values-prod.yaml \
  --set controller.image.tag=v1.1.0 \
  --set frontend.image.tag=v1.1.0
```

## Configuration

Refer to `values.yaml` for all configurable options. Customize as needed for your environment.

TODO: Add detailed configuration options as a table, including descriptions and default values.

## Backup and Recovery

### Backup PostgreSQL Database

```bash
# Create backup
kubectl exec -n idekube-prod <postgresql-pod> -- \
  pg_dump -U idekube idekube | gzip > idekube-backup-$(date +%Y%m%d).sql.gz

# Or use a CronJob for automated backups
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgresql-backup
  namespace: idekube-prod
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16-alpine
            command:
            - /bin/sh
            - -c
            - pg_dump -h idekube-postgresql -U idekube idekube | gzip > /backup/backup-$(date +%Y%m%d).sql.gz
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgresql-secret
                  key: postgres-password
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: postgresql-backup
          restartPolicy: OnFailure
```

### Restore Database

```bash
# Restore from backup
gunzip -c idekube-backup-20260104.sql.gz | \
  kubectl exec -i -n idekube-prod <postgresql-pod> -- \
  psql -U idekube idekube
```

### Backup RabbitMQ Configuration

```bash
# Export RabbitMQ definitions (users, vhosts, policies, queues)
kubectl port-forward -n idekube-prod svc/idekube-rabbitmq 15672:15672 &
curl -u idekube:$RABBITMQ_PASSWORD http://localhost:15672/api/definitions \
  > rabbitmq-definitions-$(date +%Y%m%d).json
```

### Backup secrets

```bash
# Export Kubernetes secrets
kubectl get secret postgresql-secret -n idekube-prod -o yaml > postgresql
kubectl get secret rabbitmq-secret -n idekube-prod -o yaml > rabbitmq
kubectl get secret idekube-secrets -n idekube-prod -o yaml > idekube-secrets
```

## Contributing

For bug reports and feature requests, please create an issue in the repository.

## License

[Add your license information here]
