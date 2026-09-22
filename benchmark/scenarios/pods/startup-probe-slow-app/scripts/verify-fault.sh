#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    restarts="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.restartCount}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$restarts" | grep -qE '^[1-9][0-9]*$'; then
        echo "the slow-starting container was killed and restarted"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the slow-starting container to be restarted" >&2
exit 1
