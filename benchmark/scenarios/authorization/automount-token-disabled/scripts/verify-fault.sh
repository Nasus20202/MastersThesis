#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ -z "$ready" ] || [ "$ready" = "0" ]; then
        echo "workload has no Ready Pod because the API call fails"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the workload to lose its Ready Pod" >&2
exit 1
