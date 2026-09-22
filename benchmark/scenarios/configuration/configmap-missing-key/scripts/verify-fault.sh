#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    reasons="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.state.waiting.reason}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$reasons" | grep -qx CreateContainerConfigError; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for CreateContainerConfigError" >&2
exit 1
