# idekube Helm Chart

## Testing Guide

### Prerequisites

1. Kubernetes cluster (v1.23+)
2. Helm 3.x
3. kubectl configured and connected to cluster

### Installation Steps

#### 1. Add Dependency Repositories

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

#### 2. Validate Chart Syntax

```bash
# Execute in helm/idekube directory
helm lint .
```

#### 3. Render Templates (Dry Run)

```bash
# View Kubernetes resources that will be generated
helm template idekube . --debug
```

#### 4. Install to Test Namespace

```bash
# Create test namespace
kubectl create namespace idekube-test

# Install chart
helm install idekube . \
  --namespace idekube-test \
  --set postgresql.auth.password=testpassword123 \
  --set rabbitmq.auth.password=testpassword123 \
  --set secrets.DATABASE_PASSWORD=testpassword123 \
  --set secrets.RABBITMQ_PASSWORD=testpassword123 \
  --set secrets.JWT_SECRET=test-jwt-secret-key \
  --set secrets.ENCRYPTION_KEY=test-encryption-key \
  --set ingress.enabled=false
```

#### 5. Check Deployment Status

```bash
# View release status
helm status idekube -n idekube-test

# View all pods
kubectl get pods -n idekube-test

# View all services
kubectl get svc -n idekube-test

# View detailed information
kubectl describe deployment -n idekube-test
```

#### 6. View Logs

```bash
# Controller logs
kubectl logs -n idekube-test -l app.kubernetes.io/component=controller

# RBAC logs
kubectl logs -n idekube-test -l app.kubernetes.io/component=rbac

# Housekeeping logs
kubectl logs -n idekube-test -l app.kubernetes.io/component=housekeeping

# Frontend logs
kubectl logs -n idekube-test -l app.kubernetes.io/component=frontend
```

#### 7. Test Service Connections

```bash
# Test PostgreSQL connection
kubectl port-forward -n idekube-test svc/idekube-postgresql 5432:5432
# In another terminal: psql -h localhost -U idekube -d idekube

# Test RabbitMQ management interface
kubectl port-forward -n idekube-test svc/idekube-rabbitmq 15672:15672
# Access http://localhost:15672 (username: idekube, password: testpassword123)

# Test Frontend
kubectl port-forward -n idekube-test svc/idekube-frontend 8080:80
# Access http://localhost:8080
```

#### 8. Update Configuration

```bash
# Upgrade after modifying values.yaml
helm upgrade idekube . -n idekube-test

# Or use --set parameter
helm upgrade idekube . -n idekube-test \
  --set frontend.replicaCount=3
```

#### 9. Cleanup

```bash
# Uninstall release
helm uninstall idekube -n idekube-test

# Delete namespace
kubectl delete namespace idekube-test
```

### Production Deployment Recommendations

#### 1. Prepare Production Configuration File

Create `values-prod.yaml`:

```yaml
postgresql:
  auth:
    password: "<strong-password>"
  primary:
    persistence:
      size: 100Gi

rabbitmq:
  auth:
    password: "<strong-password>"
  persistence:
    size: 20Gi

secrets:
  DATABASE_PASSWORD: "<strong-password>"
  RABBITMQ_PASSWORD: "<strong-password>"
  JWT_SECRET: "<strong-jwt-secret>"
  ENCRYPTION_KEY: "<strong-encryption-key>"

ingress:
  enabled: true
  hosts:
    - host: idekube.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
          backend:
            service:
              name: frontend
              port: 80

frontend:
  replicaCount: 3
  autoscaling:
    enabled: true

controller:
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
```

#### 2. Production Installation

```bash
helm install idekube . \
  --namespace idekube-prod \
  --create-namespace \
  --values values-prod.yaml
```

### Troubleshooting

1. **Pods Cannot Start**
   ```bash
   kubectl describe pod <pod-name> -n idekube-test
   kubectl logs <pod-name> -n idekube-test
   ```

2. **ImagePullBackOff Error**
   - Check if image registry is accessible
   - Verify imagePullSecrets configuration is correct

3. **Database Connection Failed**
   - Verify PostgreSQL pod is running
   - Check connection string and password
   - View database logs

4. **Inter-service Communication Failed**
   - Check Service DNS resolution
   - Verify RabbitMQ is accessible
   - Review network policy configuration

### Custom Configuration Examples

#### Enable Ingress with TLS

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: idekube.example.com
      paths:
        - path: /
          pathType: Prefix
          backend:
            service:
              name: frontend
              port: 80
  tls:
    - secretName: idekube-tls
      hosts:
        - idekube.example.com
```

#### Configure Resource Limits

```yaml
controller:
  resources:
    limits:
      cpu: 2000m
      memory: 2Gi
    requests:
      cpu: 500m
      memory: 512Mi
```

#### Configure Persistent Storage

```yaml
postgresql:
  primary:
    persistence:
      enabled: true
      size: 50Gi
      storageClass: fast-ssd
```

### Monitoring and Alerting

Integrate Prometheus and Grafana:

```yaml
postgresql:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true

rabbitmq:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
```

### Backup and Recovery

```bash
# Backup PostgreSQL data
kubectl exec -n idekube-prod <postgresql-pod> -- pg_dump -U idekube idekube > backup.sql

# Restore data
kubectl exec -i -n idekube-prod <postgresql-pod> -- psql -U idekube idekube < backup.sql
```
