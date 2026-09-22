#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    desired="$(kubectl get deployment app -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    available="$(kubectl get deployment app -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true)"
    if [ -n "$desired" ] && [ "$available" = "$desired" ]; then
        printf 'desired=%s available=%s\n' "$desired" "$available"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for all Deployment replicas to be available" >&2
exit 1
