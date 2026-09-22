#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    desired="$(kubectl get daemonset app -o jsonpath='{.status.desiredNumberScheduled}' 2>/dev/null || true)"
    ready="$(kubectl get daemonset app -o jsonpath='{.status.numberReady}' 2>/dev/null || true)"
    if [ -z "$desired" ]; then
        desired=0
    fi
    if [ -z "$ready" ]; then
        ready=0
    fi
    if [ "$desired" -ge 1 ] && [ "$ready" = "$desired" ]; then
        printf 'desired=%s ready=%s\n' "$desired" "$ready"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for every desired DaemonSet Pod to be ready" >&2
exit 1
