#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    status="$(kubectl get gateway app-gateway -o jsonpath='{.status.conditions[?(@.type=="Programmed")].status}' 2>/dev/null || true)"
    if [ "$status" = "True" ]; then
        printf 'gateway app-gateway is Programmed\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for gateway app-gateway to be Programmed" >&2
exit 1
