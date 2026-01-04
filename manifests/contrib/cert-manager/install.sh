#!/bin/bash

set -e

NAMESPACE="cert-manager"
CERT_MANAGER_VERSION="${CERT_MANAGER_VERSION:-v1.19.2}"

echo "Installing cert-manager ${CERT_MANAGER_VERSION}..."

# Add cert-manager Helm repository
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Create namespace
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Install cert-manager with CRDs
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace ${NAMESPACE} \
  --version ${CERT_MANAGER_VERSION} \
  --set installCRDs=true \
  --wait

echo "Waiting for cert-manager to be ready..."
kubectl wait --for=condition=ready pod -l app=cert-manager -n ${NAMESPACE} --timeout=300s

echo "cert-manager installation complete!"
echo ""
echo "Next steps:"
echo "  1. Create Cloudflare API secret: kubectl apply -f cloudflare-api-secret.yaml"
echo "  2. Create ClusterIssuer: kubectl apply -f cloudflare-cluster-issuer.yaml"
echo ""
echo "To verify installation:"
echo "  kubectl get pods -n ${NAMESPACE}"
