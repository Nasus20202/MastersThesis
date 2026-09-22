#!/bin/sh
set -eu

kubectl patch configmap app-config --type=merge -p '{"immutable":true}'
kubectl patch configmap app-config --type=merge -p '{"data":{"message":"v2"}}' || true
