#!/bin/sh
set -eu

kubectl get namespace legacy-app >/dev/null

if kubectl get deployment app -n legacy-app >/dev/null 2>&1; then
    echo "namespace legacy-app already contains a Deployment" >&2
    exit 1
fi

printf 'namespace legacy-app is empty and ready\n'
