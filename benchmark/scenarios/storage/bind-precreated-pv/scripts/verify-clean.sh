#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

phase=""
while [ "$(date +%s)" -lt "$deadline" ]; do
    phase="$(kubectl get pv app-pv -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    if [ "$phase" = "Available" ]; then
        break
    fi
    sleep 1
done
printf 'pv_phase=%s\n' "$phase"
test "$phase" = "Available"

if kubectl get pvc app-pvc >/dev/null 2>&1; then
    echo "pvc app-pvc unexpectedly exists before the task" >&2
    exit 1
fi
