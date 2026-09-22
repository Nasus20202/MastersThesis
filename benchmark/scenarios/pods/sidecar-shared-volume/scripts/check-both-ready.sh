#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    count="$(kubectl get pod app -o jsonpath='{.spec.containers[*].name}' 2>/dev/null | wc -w | tr -d ' ' || true)"
    statuses="$(kubectl get pod app -o jsonpath='{range .status.containerStatuses[*]}{.ready}{"\n"}{end}' 2>/dev/null || true)"
    if [ "${count:-0}" = "2" ] && [ -n "$statuses" ] && ! printf '%s\n' "$statuses" | grep -qv '^true$'; then
        echo "both containers in the Pod are Ready"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for both containers to be Ready" >&2
exit 1
