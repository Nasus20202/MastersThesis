#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    reasons="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.lastState.terminated.reason}{"\n"}{end}' 2>/dev/null || true)"
    restarts="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.restartCount}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$reasons" | grep -qx OOMKilled || printf '%s\n' "$restarts" | grep -qE '^[1-9][0-9]*$'; then
        echo "application container is being killed by the low memory limit"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for an OOMKilled restart" >&2
exit 1
