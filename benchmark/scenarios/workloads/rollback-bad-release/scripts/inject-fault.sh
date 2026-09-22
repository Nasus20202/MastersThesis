#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","image":"busybox:1.37.0","command":["sh","-c","exit 1"]}]}}}}'

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    available="$(kubectl get deployment app -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true)"
    if [ -z "$available" ]; then
        available=0
    fi
    if [ "$available" -lt 3 ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the bad release to reduce the available replicas" >&2
exit 1
