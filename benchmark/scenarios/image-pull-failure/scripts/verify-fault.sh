#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    reason="$(kubectl get pod -l app=app -o jsonpath='{.items[0].status.containerStatuses[0].state.waiting.reason}' 2>/dev/null || true)"
    if [ "$reason" = "ImagePullBackOff" ]; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for ImagePullBackOff" >&2
exit 1
