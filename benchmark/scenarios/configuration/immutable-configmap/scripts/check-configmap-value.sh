#!/bin/sh
set -eu

expected="${1:?expected value is required}"
timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    message="$(kubectl get configmap app-config -o jsonpath='{.data.message}' 2>/dev/null || true)"
    if [ "$message" = "$expected" ]; then
        printf 'configmap_value=%s\n' "$message"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the ConfigMap to hold the expected value" >&2
exit 1
