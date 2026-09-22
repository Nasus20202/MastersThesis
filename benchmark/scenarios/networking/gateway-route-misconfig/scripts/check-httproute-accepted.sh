#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    parent="$(kubectl get httproute app-route -o jsonpath='{.spec.parentRefs[0].name}' 2>/dev/null || true)"
    if [ "$parent" = "app-gateway" ] && kubectl get gateway app-gateway >/dev/null 2>&1; then
        status="$(kubectl get httproute app-route -o jsonpath='{.status.parents[*].conditions[?(@.type=="Accepted")].status}' 2>/dev/null || true)"
        if printf '%s\n' "$status" | tr ' ' '\n' | grep -qx True; then
            printf 'httproute app-route is Accepted by gateway app-gateway\n'
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for httproute app-route to be Accepted by gateway app-gateway" >&2
exit 1
