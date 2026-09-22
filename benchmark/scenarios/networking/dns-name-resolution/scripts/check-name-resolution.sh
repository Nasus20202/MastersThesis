#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))
consecutive=0

while [ "$(date +%s)" -lt "$deadline" ]; do
    pod="$(kubectl get pods -l app=app -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
    if [ -n "$pod" ]; then
        status="$(kubectl exec "$pod" -- cat /shared/status 2>/dev/null || true)"
        if [ "$status" = "ok" ]; then
            consecutive=$((consecutive + 1))
            if [ "$consecutive" -ge 3 ]; then
                printf 'client resolved and reached the peer for %s consecutive checks\n' "$consecutive"
                exit 0
            fi
        else
            consecutive=0
        fi
    fi
    sleep 2
done

echo "timed out waiting for the application to resolve and reach its peer Service" >&2
exit 1
