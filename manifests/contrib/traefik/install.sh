#!/bin/bash

set -e

NAMESPACE="traefik"
TRAEFIK_VERSION="${TRAEFIK_VERSION:-38.0.1}"

echo "Installing Traefik ${TRAEFIK_VERSION}..."

# # Add Traefik Helm repository
# helm repo add traefik https://traefik.github.io/charts
# helm repo update

# Create namespace
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Install Traefik
helm upgrade --install traefik traefik \
  --repo https://traefik.github.io/charts \
  --namespace ${NAMESPACE} \
  --version ${TRAEFIK_VERSION} \
  --set "ports.websecure.tls.enabled=true" \
  --set "ingressRoute.dashboard.enabled=true" \
  --wait
  # --set "deployment.kind=DaemonSet" \
  # --set "hostNetwork=true" \
  # --set "updateStrategy.rollingUpdate.maxUnavailable=1" \
  # --set "updateStrategy.rollingUpdate.maxSurge=0" \
  # --wait

echo "Traefik installation complete!"
echo ""
echo "To access Traefik dashboard, run:"
echo "  kubectl port-forward -n ${NAMESPACE} \$(kubectl get pods -n ${NAMESPACE} --selector 'app.kubernetes.io/name=traefik' --output=name) 9000:9000"
echo "  Then open http://localhost:9000/dashboard/"
