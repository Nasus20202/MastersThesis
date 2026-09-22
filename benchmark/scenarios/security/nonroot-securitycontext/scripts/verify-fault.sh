#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    detail="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.state.waiting.reason} {.state.waiting.message}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$detail" | grep -q 'runAsNonRoot'; then
        echo "container is rejected because runAsNonRoot conflicts with the image user"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the container to be rejected by runAsNonRoot" >&2
exit 1
