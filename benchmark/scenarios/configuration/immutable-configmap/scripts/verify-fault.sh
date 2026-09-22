#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    immutable="$(kubectl get configmap app-config -o jsonpath='{.immutable}' 2>/dev/null || true)"
    message="$(kubectl get configmap app-config -o jsonpath='{.data.message}' 2>/dev/null || true)"
    if [ "$immutable" = "true" ] && [ "$message" = "v1" ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the ConfigMap to remain immutable at v1" >&2
exit 1
