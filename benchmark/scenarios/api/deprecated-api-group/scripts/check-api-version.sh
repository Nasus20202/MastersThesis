#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    api_version="$(kubectl get deployment app -n legacy-app -o jsonpath='{.apiVersion}' 2>/dev/null || true)"
    if [ "$api_version" = "apps/v1" ]; then
        printf 'deployment_api_version=%s\n' "$api_version"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for a Deployment served by apps/v1" >&2
exit 1
