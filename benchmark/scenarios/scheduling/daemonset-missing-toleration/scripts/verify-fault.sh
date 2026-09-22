#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    desired="$(kubectl get daemonset app -o jsonpath='{.status.desiredNumberScheduled}' 2>/dev/null || true)"
    if [ -z "$desired" ]; then
        desired=0
    fi
    if [ "$desired" -eq 0 ]; then
        printf 'desired=%s\n' "$desired"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the DaemonSet to no longer target the node" >&2
exit 1
