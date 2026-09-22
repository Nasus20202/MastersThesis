#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    class="$(kubectl get pvc data-app-0 -o jsonpath='{.spec.storageClassName}' 2>/dev/null || true)"
    phase="$(kubectl get pvc data-app-0 -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    ready="$(kubectl get statefulset app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ "$class" = "fast-ssd" ] && [ "$phase" = "Pending" ] && [ "${ready:-0}" != "2" ]; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the volume claim template to stay Pending" >&2
exit 1
