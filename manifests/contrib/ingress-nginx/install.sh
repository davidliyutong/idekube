#!/bin/bash

set -e

NAMESPACE="ingress-nginx"
CHART_VERSION="${CHART_VERSION:-4.14.1}"
APP_VERSION="${APP_VERSION:-1.14.1}"

echo "Installing Ingress NGINX ${CHART_VERSION}..."

# Add Ingress NGINX Helm repository
# helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
# helm repo update

# Create namespace
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Install Ingress NGINX
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ${NAMESPACE} \
  --version ${CHART_VERSION} \
  --set controller.image.digest="" \
  --set controller.image.digestChroot="" \
  --set controller.admissionWebhooks.patch.image.digest="" \
  --set controller.dnsPolicy="ClusterFirstWithHostNet" \
  --set controller.hostNetwork=true \
  --set controller.publishService.enabled=false \
  --set controller.kind=DaemonSet \
  --set controller.service.enabled=false \
  --wait

echo "Ingress NGINX installation complete!"