#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    requests="$(kubectl get pods -l app=app -o jsonpath='{.items[0].spec.containers[0].resources.requests.cpu}' 2>/dev/null || true)"
    limits="$(kubectl get pods -l app=app -o jsonpath='{.items[0].spec.containers[0].resources.limits.cpu}' 2>/dev/null || true)"
    if [ "$requests" = "100m" ] && [ "$limits" = "200m" ]; then
        printf 'requests=%s limits=%s\n' "$requests" "$limits"
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the LimitRange defaults to be applied" >&2
exit 1
