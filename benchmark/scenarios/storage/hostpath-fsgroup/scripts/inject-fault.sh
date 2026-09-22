#!/bin/sh
set -eu

kubectl create -f manifests/break-perms.yaml
kubectl wait --for=condition=complete job/data-perms-break --timeout=60s
kubectl rollout restart deployment/app
