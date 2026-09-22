#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    pending="$(kubectl get pods -l app=app --field-selector=status.phase=Pending -o name 2>/dev/null || true)"
    updated="$(kubectl get deployment app -o jsonpath='{.status.updatedReplicas}' 2>/dev/null || true)"
    if [ -n "$pending" ] && { [ -z "$updated" ] || [ "$updated" != "3" ]; }; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for Pending Pods caused by the node selector" >&2
exit 1
