#!/bin/sh
set -eu

timeout_seconds=90
settle_seconds=8
expected_replicas=1
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$ready" = "$expected_replicas" ]; then
        sleep "$settle_seconds"
        ready_after="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
        if [ "$ready_after" = "$expected_replicas" ]; then
            printf 'ready_replicas=%s\n' "$ready_after"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for the application Pod to be Ready" >&2
exit 1
