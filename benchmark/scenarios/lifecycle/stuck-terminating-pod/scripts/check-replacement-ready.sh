#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    deleting="$(kubectl get pod app -o jsonpath='{.metadata.deletionTimestamp}' 2>/dev/null || true)"
    phase="$(kubectl get pod app -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    ready="$(kubectl get pod app -o jsonpath='{range .status.conditions[?(@.type=="Ready")]}{.status}{end}' 2>/dev/null || true)"
    if [ -z "$deleting" ] && [ "$phase" = "Running" ] && [ "$ready" = "True" ]; then
        echo "a healthy replacement Pod named app is Running and Ready"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for a healthy replacement Pod named app" >&2
exit 1
