#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    restarts="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.restartCount}{"\n"}{end}' 2>/dev/null || true)"
    crashing=false
    for count in $restarts; do
        if [ "$count" -ge 1 ]; then
            crashing=true
        fi
    done
    if [ "$crashing" = true ] && [ "${ready:-0}" != "3" ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the hostPath ownership change to break the container" >&2
exit 1
