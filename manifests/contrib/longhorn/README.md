# Longhorn

Longhorn is a lightweight, reliable, and powerful distributed block storage system for Kubernetes.

## Installation

```bash
./install.sh
```

### Specify Version

Set the `LONGHORN_VERSION` environment variable:

```bash
LONGHORN_VERSION=v1.10.1 ./install.sh
```

## Resources

- **Official Website:** <https://longhorn.io/>
- **Documentation:** <https://longhorn.io/docs/>
- **Releases:** <https://github.com/longhorn/longhorn/releases>
- **GitHub:** <https://github.com/longhorn/longhorn>

## Access UI

After installation, access the Longhorn UI:

```bash
kubectl port-forward -n longhorn-system svc/longhorn-frontend 8080:80
```

Then open <http://localhost:8080>
