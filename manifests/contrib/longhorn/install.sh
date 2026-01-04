#!/bin/bash

set -e

NAMESPACE="longhorn-system"
LONGHORN_VERSION="${LONGHORN_VERSION:-v1.10.1}"

echo "Installing Longhorn ${LONGHORN_VERSION}..."

# Create namespace
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Install Longhorn using kubectl
kubectl apply -n ${NAMESPACE} -f https://raw.githubusercontent.com/longhorn/longhorn/${LONGHORN_VERSION}/deploy/longhorn.yaml

echo "Waiting for Longhorn to be ready..."
kubectl wait --for=condition=ready pod -l app=longhorn-manager -n ${NAMESPACE} --timeout=300s

echo "Longhorn installation complete!"
echo ""
echo "To access Longhorn UI, run:"
echo "  kubectl port-forward -n ${NAMESPACE} svc/longhorn-frontend 8080:80"
echo "  Then open http://localhost:8080"
