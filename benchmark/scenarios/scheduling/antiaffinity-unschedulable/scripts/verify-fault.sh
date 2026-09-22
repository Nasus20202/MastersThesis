#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    conditions="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.conditions[?(@.type=="PodScheduled")]}{.status}:{.reason}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$conditions" | grep -q '^False:Unschedulable$'; then
        echo "required pod anti-affinity leaves a replica unschedulable"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for a replica blocked by required pod anti-affinity" >&2
exit 1
