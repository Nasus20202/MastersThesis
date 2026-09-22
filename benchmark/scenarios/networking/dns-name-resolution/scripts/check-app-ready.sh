#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))
expected="${1:-2}"

while [ "$(date +%s)" -lt "$deadline" ]; do
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "${ready:-0}" = "$expected" ]; then
        printf 'ready_replicas=%s\n' "$ready"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for $expected Ready application replicas" >&2
exit 1
