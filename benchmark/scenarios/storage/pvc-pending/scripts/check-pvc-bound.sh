#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    phase="$(kubectl get pvc app-pvc -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    if [ "$phase" = "Bound" ]; then
        printf 'pvc_phase=%s\n' "$phase"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for app-pvc to become Bound" >&2
exit 1
