#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    phase="$(kubectl get pvc app-pvc -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    class="$(kubectl get pvc app-pvc -o jsonpath='{.spec.storageClassName}' 2>/dev/null || true)"
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$phase" = "Pending" ] && [ "$class" = "fast-ssd" ] && [ "${ready:-0}" != "3" ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the claim to stay Pending" >&2
exit 1
