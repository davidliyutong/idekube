# Quick Reference: Environment Configurations

| Field                | Dev                        | Prod                   |
| -------------------- | -------------------------- | ---------------------- |
| **Namespace**        | `angos-dev`                | `angos-prod`           |
| **Ingress Hostname** | `registry-dev.example.com` | `registry.example.com` |
| **Ingress Class**    | `traefik`                  | `nginx`                |
| **Cert Issuer**      | `letsencrypt-staging`      | `letsencrypt-prod`     |
| **Storage Size**     | `10Gi`                     | `100Gi`                |
| **Storage Class**    | `standard`                 | `fast-ssd`             |

## Password Hash Generation

To generate a new Argon2 password hash for the registry authentication, use the following command:

```bash
echo -n "yourpassword" | argon2 somesalt -e
```

## Quick Commands

```bash
# Deploy Dev
kubectl apply -k manifests/contrib/angos/overlays/dev

# Deploy Prod
kubectl apply -k manifests/contrib/angos/overlays/prod

# Preview without applying
kubectl kustomize manifests/contrib/angos/overlays/dev

# Diff between environments
diff <(kubectl kustomize manifests/contrib/angos/overlays/dev) \
     <(kubectl kustomize manifests/contrib/angos/overlays/prod)

# Delete deployment
kubectl delete -k manifests/contrib/angos/overlays/dev
```

## Customization Checklist

Before deploying:

- [ ] Update ingress hostname in `ingress-patch.yaml`
- [ ] Set correct ingress class for your cluster
- [ ] Configure cert-manager issuer name
- [ ] Adjust PVC size based on needs
- [ ] Set appropriate storage class
- [ ] Generate new htpasswd hash
- [ ] Change registry-secret to a secure value
- [ ] Update namespace if needed
