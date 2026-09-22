#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    states="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.restartCount}:{.lastState.terminated.exitCode}:{.state.waiting.reason}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$states" | grep -Eq '^[1-9][0-9]*:[1-9][0-9]*:|:CrashLoopBackOff$'; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for CrashLoopBackOff" >&2
exit 1
