#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    running="$(kubectl get pods -l app=app -o jsonpath='{range .items[?(@.status.phase=="Running")]}{.metadata.name}{"\n"}{end}' 2>/dev/null || true)"
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ -n "$running" ] && { [ -z "$ready" ] || [ "$ready" = "0" ]; }; then
        printf 'running_pods=%s ready_replicas=%s\n' "$(printf '%s\n' "$running" | grep -c .)" "${ready:-0}"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for running Pods that are not Ready" >&2
exit 1
