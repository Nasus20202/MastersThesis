#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    replicas="$(kubectl get statefulset app -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    ready="$(kubectl get statefulset app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$replicas" = "2" ] && [ "$ready" = "2" ]; then
        printf 'replicas=%s ready=%s\n' "$replicas" "$ready"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for two ready StatefulSet Pods" >&2
exit 1
