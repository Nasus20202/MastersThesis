#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    for pod in $(kubectl get pods -l app=app -o name 2>/dev/null || true); do
        if kubectl exec "$pod" -- sh -c 'echo ok > /data/write-check' >/dev/null 2>&1; then
            printf 'volume writable in %s\n' "$pod"
            exit 0
        fi
    done
    sleep 2
done

echo "timed out waiting for a writable hostPath volume" >&2
exit 1
