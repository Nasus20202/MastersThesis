#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    conditions="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.conditions[?(@.type=="PodScheduled")]}{.status}:{.reason}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$conditions" | grep -q '^False:Unschedulable$'; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for an unschedulable application Pod" >&2
exit 1
