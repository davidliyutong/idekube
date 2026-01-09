#!/bin/bash

set -e

LONGHORN_VERSION="${LONGHORN_VERSION:-v1.10.1}"

kubectl create -f https://raw.githubusercontent.com/longhorn/longhorn/${LONGHORN_VERSION}/uninstall/uninstall.yaml
kubectl get job/longhorn-uninstall -n longhorn-system -w

kubectl delete namespace longhorn-system --force --grace-period=0