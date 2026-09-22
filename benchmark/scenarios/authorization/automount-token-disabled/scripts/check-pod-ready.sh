#!/bin/sh
set -eu

timeout_seconds=150
stable_seconds=20
deadline=$(( $(date +%s) + timeout_seconds ))

ready_now() {
    replicas="$(kubectl get deployment app -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    [ "$replicas" = "1" ] && [ "$ready" = "1" ]
}

while [ "$(date +%s)" -lt "$deadline" ]; do
    if ready_now; then
        sleep "$stable_seconds"
        if ready_now; then
            printf 'replicas=1 ready=1 for %ss\n' "$stable_seconds"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for a stable Ready Pod" >&2
exit 1
