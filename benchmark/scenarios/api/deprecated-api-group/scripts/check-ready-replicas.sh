#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    replicas="$(kubectl get deployment app -n legacy-app -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    updated="$(kubectl get deployment app -n legacy-app -o jsonpath='{.status.updatedReplicas}' 2>/dev/null || true)"
    ready="$(kubectl get deployment app -n legacy-app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$replicas" = "3" ] && [ "$updated" = "3" ] && [ "$ready" = "3" ]; then
        printf 'replicas=%s updated=%s ready=%s\n' "$replicas" "$updated" "$ready"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for three Ready replicas" >&2
exit 1
