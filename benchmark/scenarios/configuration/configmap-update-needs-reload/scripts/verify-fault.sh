#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    message="$(kubectl get configmap app-config -o jsonpath='{.data.message}' 2>/dev/null || true)"
    stale=true
    count=0
    for pod in $(kubectl get pods -l app=app -o name 2>/dev/null || true); do
        value="$(kubectl exec "$pod" -- printenv CONFIG_MESSAGE 2>/dev/null || true)"
        if [ "$value" != "alpha" ]; then
            stale=false
        fi
        count=$(( count + 1 ))
    done
    if [ "$message" = "beta" ] && [ "$stale" = true ] && [ "$count" -ge 1 ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the running Pods to keep the old configuration value" >&2
exit 1
