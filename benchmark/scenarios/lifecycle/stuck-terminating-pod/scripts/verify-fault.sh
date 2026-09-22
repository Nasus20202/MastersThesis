#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    deleting="$(kubectl get pod app -o jsonpath='{.metadata.deletionTimestamp}' 2>/dev/null || true)"
    if [ -n "$deleting" ]; then
        echo "Pod app is stuck Terminating"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for Pod app to enter Terminating" >&2
exit 1
