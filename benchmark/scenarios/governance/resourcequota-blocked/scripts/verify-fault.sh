#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    failure="$(kubectl get deployment app -o jsonpath='{.status.conditions[?(@.type=="ReplicaFailure")].status}' 2>/dev/null || true)"
    desired="$(kubectl get deployment app -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$failure" = "True" ] && [ "$desired" = "3" ] && [ "${ready:-0}" != "3" ]; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the ResourceQuota to block a replica" >&2
exit 1
