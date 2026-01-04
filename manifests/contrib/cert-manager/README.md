# cert-manager

cert-manager adds certificates and certificate issuers as resource types in Kubernetes clusters, and simplifies the process of obtaining, renewing and using those certificates.

## Pre-requisites

- Kubernetes >= v1.25
- `open-iscsi` (iscsi-initiator-utils)
- `cryptsetup` if using encrypted volumes
- `nfs-common`, if using NFS for storage class
- `bash`, `curl`, `findmnt`, `grep`, `awk`, `blkid`, `lsblk`, `jq` installed on the machine running the script

## Installation

```bash
./install.sh
```

### Specify Version

Set the `CERT_MANAGER_VERSION` environment variable (Helm chart version):

```bash
CERT_MANAGER_VERSION=v1.19.2 ./install.sh
```

## Cloudflare DNS-01 Challenge Setup

1. **Create Cloudflare API Token Secret:**

   Edit `cloudflare-api-secret.yaml` with your Cloudflare API token, then:

   ```bash
   kubectl apply -f cloudflare-api-secret.yaml
   ```

2. **Create ClusterIssuer:**

   Edit `cloudflare-cluster-issuer.yaml` with your email, then:

   ```bash
   kubectl apply -f cloudflare-cluster-issuer.yaml
   ```

### Get Cloudflare API Token

1. Go to <https://dash.cloudflare.com/profile/api-tokens>
2. Create a token with:
   - **Permissions:** `Zone:DNS:Edit` and `Zone:Zone:Read`
   - **Zone Resources:** Include the zones you want to manage

## Resources

- **Official Website:** <https://cert-manager.io/>
- **Documentation:** <https://cert-manager.io/docs/>
- **Helm Chart Releases:** <https://github.com/cert-manager/cert-manager/releases>
- **GitHub:** <https://github.com/cert-manager/cert-manager>

## Usage Example

Create a Certificate resource:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com
  namespace: default
spec:
  secretName: example-com-tls
  issuerRef:
    name: letsencrypt-cloudflare
    kind: ClusterIssuer
  dnsNames:
  - example.com
  - "*.example.com"
```
