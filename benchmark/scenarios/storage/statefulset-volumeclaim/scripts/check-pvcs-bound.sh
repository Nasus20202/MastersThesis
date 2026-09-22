#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    count=0
    bad=false
    for claim in data-app-0 data-app-1; do
        phase="$(kubectl get pvc "$claim" -o jsonpath='{.status.phase}' 2>/dev/null || true)"
        if [ -n "$phase" ]; then
            count=$(( count + 1 ))
            if [ "$phase" != "Bound" ]; then
                bad=true
            fi
        fi
    done
    if [ "$count" -ge 1 ] && [ "$bad" = false ]; then
        printf 'bound_claims=%s\n' "$count"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the StatefulSet claims to bind" >&2
exit 1
