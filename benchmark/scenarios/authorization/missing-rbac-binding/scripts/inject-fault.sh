#!/bin/sh
set -eu

kubectl patch rolebinding app-config-reader --type=json \
    -p='[{"op":"replace","path":"/subjects/0/name","value":"wrong-reader"}]'
kubectl rollout restart deployment/app
