#!/bin/bash

set -e
kubectl create -f https://raw.githubusercontent.com/longhorn/longhorn/v1.10.1/uninstall/uninstall.yaml
kubectl get job/longhorn-uninstall -n longhorn-system -w

kubectl delete namespace longhorn-system --force --grace-period=0