# Traefik

Traefik is a modern HTTP reverse proxy and load balancer that makes deploying microservices easy.

## Installation

```bash
./install.sh
```

### Specify Version

Set the `TRAEFIK_VERSION` environment variable (Helm chart version):

```bash
TRAEFIK_VERSION=v38.0.1 ./install.sh
```

## Resources

- **Official Website:** <https://traefik.io/>
- **Documentation:** <https://doc.traefik.io/traefik/>
- **Helm Chart Releases:** <https://github.com/traefik/traefik-helm-chart/releases>
- **GitHub:** <https://github.com/traefik/traefik>

## Access Dashboard

After installation, access the Traefik dashboard:

```bash
kubectl port-forward -n traefik $(kubectl get pods -n traefik --selector 'app.kubernetes.io/name=traefik' --output=name) 9000:9000
```

Then open <http://localhost:9000/dashboard/>