#!/bin/sh
set -eu

timeout_seconds=120
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    reasons="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.state.waiting.reason}{"\n"}{end}' 2>/dev/null || true)"
    terminates="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.lastState.terminated.exitCode}{"\n"}{end}' 2>/dev/null || true)"
    restarts="$(kubectl get pods -l app=app -o jsonpath='{range .items[*].status.containerStatuses[*]}{.restartCount}{"\n"}{end}' 2>/dev/null || true)"
    if printf '%s\n' "$reasons" | grep -Eq 'CrashLoopBackOff|RunContainerError|Error'; then
        exit 0
    fi
    for code in $terminates; do
        if [ "$code" != "0" ]; then
            exit 0
        fi
    done
    for count in $restarts; do
        if [ "$count" -ge 1 ]; then
            exit 0
        fi
    done
    sleep 2
done

echo "timed out waiting for the read-only root filesystem to break the container" >&2
exit 1
