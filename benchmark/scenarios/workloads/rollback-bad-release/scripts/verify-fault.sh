#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    image="$(kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || true)"
    available="$(kubectl get deployment app -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true)"
    if [ -z "$available" ]; then
        available=0
    fi
    if [ "$image" = "busybox:1.37.0" ] && [ "$available" -lt 3 ]; then
        printf 'current_image=%s available=%s\n' "$image" "$available"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the bad release to be the current revision" >&2
exit 1
