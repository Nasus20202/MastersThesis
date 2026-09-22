#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    stuck="$(kubectl get pods -o jsonpath='{range .items[?(@.metadata.deletionTimestamp)]}{.metadata.name}{"\n"}{end}' 2>/dev/null || true)"
    if [ -z "$stuck" ]; then
        echo "no Pod is stuck terminating"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the terminating Pod to be removed" >&2
exit 1
