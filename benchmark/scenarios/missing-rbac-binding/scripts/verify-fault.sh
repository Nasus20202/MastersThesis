#!/bin/sh
set -eu

service_account='system:serviceaccount:default:config-reader'
allowed="$(kubectl auth can-i get configmap/app-config --namespace=default --as="$service_account" || true)"
test "$allowed" = "no"

if kubectl rollout status deployment/app --timeout=15s >/dev/null 2>&1; then
    echo "deployment unexpectedly completed without required ConfigMap access" >&2
    exit 1
fi
