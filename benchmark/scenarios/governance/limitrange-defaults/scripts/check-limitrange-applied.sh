#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    pod="$(running_pod app=app)"
    if [ -n "$pod" ]; then
        requests="$(kubectl get pod "$pod" -o jsonpath='{.spec.containers[0].resources.requests.cpu}' 2>/dev/null || true)"
        limits="$(kubectl get pod "$pod" -o jsonpath='{.spec.containers[0].resources.limits.cpu}' 2>/dev/null || true)"
    else
        requests=""
        limits=""
    fi
    if [ "$requests" = "100m" ] && [ "$limits" = "200m" ]; then
        printf 'requests=%s limits=%s\n' "$requests" "$limits"
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the LimitRange defaults to be applied" >&2
exit 1
