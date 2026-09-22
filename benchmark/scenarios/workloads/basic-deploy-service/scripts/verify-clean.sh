#!/bin/sh
set -eu

kubectl get namespace basic-app >/dev/null

if kubectl get deployment,service -n basic-app -o name 2>/dev/null | grep -q .; then
    echo "namespace basic-app is not empty" >&2
    exit 1
fi

printf 'namespace basic-app is empty and ready\n'
