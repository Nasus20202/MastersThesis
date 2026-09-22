#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    phase="$(kubectl get pvc app-pvc -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    if [ "$phase" = "Bound" ]; then
        printf 'pvc_phase=%s\n' "$phase"
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the PersistentVolumeClaim to bind" >&2
exit 1
