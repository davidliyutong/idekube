# RustFS Object Storage

This directory contains Kubernetes manifests for deploying RustFS object storage server.

## Directory Structure

This configuration uses Kustomize for deployment management:

```
rustfs/
├── base/                    # Base manifests
│   ├── kustomization.yaml
│   ├── rustfs-configmap.yaml
│   ├── rustfs-ingress.yaml
│   ├── rustfs-pvc.yaml
│   ├── rustfs-secret.yaml
│   ├── rustfs-service.yaml
│   └── rustfs-statefulset.yaml
└── overlays/               # Environment-specific overlays
    └── dev/
        ├── ingress-patch.yaml
        ├── kustomization.yaml
        ├── pvc-patch.yaml
        └── secret-patch.yaml
```

## Components

- **rustfs-statefulset.yaml**: StatefulSet configuration for RustFS server
- **rustfs-service.yaml**: ClusterIP Service for API and Console
- **rustfs-secret.yaml**: Credentials for RustFS access (access key and secret key)
- **rustfs-configmap.yaml**: Configuration for RustFS server
- **rustfs-pvc.yaml**: Persistent Volume Claim for data storage
- **rustfs-ingress.yaml**: Ingress configuration for external access

## Deployment

```bash
# Create namespace if it doesn't exist
kubectl create namespace idekube

# Apply using Kustomize (base configuration)
kubectl apply -k manifests/services/rustfs/base/

# Or apply dev overlay
kubectl apply -k manifests/services/rustfs/overlays/dev/
```

## Access

### Internal Access

- **API Endpoint**: `http://rustfs.idekube.svc.cluster.local:9000`
- **Console**: `http://rustfs.idekube.svc.cluster.local:9001`

### External Access (via Ingress)

- **API Endpoint**: `https://rustfs.example.com`
- **Console**: `https://rustfs-console.example.com`

Update the host names in [base/rustfs-ingress.yaml](base/rustfs-ingress.yaml) to match your domain.

**Note**: The ingress is configured with:
- Ingress class: `nginx`
- TLS enabled with cert-manager using `letsencrypt` cluster issuer
- TLS secret name: `rustfs-tls`

### Port Forward for Local Access

```bash
# Access API
kubectl port-forward -n idekube svc/rustfs 9000:9000

# Access Console
kubectl port-forward -n idekube svc/rustfs 9001:9001
```

## Configuration

### Credentials

Default credentials (change in production!):
- **Access Key**: rustfsadmin
- **Secret Key**: rustfsadmin

Update credentials in [base/rustfs-secret.yaml](base/rustfs-secret.yaml) before deployment.

### Storage

- Default PVC size: 100Gi
- Storage class: Can be specified in PVC spec (currently commented out)
- Path: `/data` inside container
- Access Mode: ReadWriteOnce

### Resources

Default resource limits:
- CPU: 1000m (limit), 50m (request)
- Memory: 2Gi (limit), 128Mi (request)

## Scaling

RustFS is currently configured as a single-node setup. For high availability:

```bash
# Scale replicas (if RustFS supports distributed mode)
kubectl scale statefulset rustfs -n idekube --replicas=3
```

## Monitoring

Check status:

```bash
# View pods
kubectl get pods -n idekube -l app=rustfs

# View logs
kubectl logs -n idekube -l app=rustfs

# Check PVC
kubectl get pvc -n idekube
```

## Backup

Backup persistent data:

```bash
# Get PVC name
kubectl get pvc -n idekube

# Create snapshot or backup based on your storage provider
```

## Troubleshooting

1. **Pod not starting**:
   ```bash
   kubectl describe pod -n idekube -l app=rustfs
   kubectl logs -n idekube -l app=rustfs
   ```

2. **Storage issues**:
   ```bash
   kubectl get pvc -n idekube
   kubectl describe pvc -n idekube
   ```

3. **Connection issues**:
   ```bash
   kubectl get svc -n idekube rustfs
   kubectl exec -n idekube -it rustfs-0 -- sh
   ```
