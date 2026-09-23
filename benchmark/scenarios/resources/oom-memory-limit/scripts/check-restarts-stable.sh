#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

max_restarts() {
    restart_counts app=app | tr -s ' ' '\n' | grep -E '^[0-9]+$' | sort -n | tail -n 1
}

ready_replicas=''
while [ "$(date +%s)" -lt "$deadline" ]; do
    ready_replicas="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ -n "$ready_replicas" ] && [ "$ready_replicas" != "0" ]; then
        break
    fi
    sleep 2
done

if [ -z "$ready_replicas" ] || [ "$ready_replicas" = "0" ]; then
    echo "application Pods never became Ready" >&2
    exit 1
fi

markers="$(container_markers app=app)"
if printf '%s\n' "$markers" | grep -qE 'OOMKilled|CrashLoopBackOff|Error'; then
    echo "an application container is still being killed" >&2
    exit 1
fi

before="$(max_restarts)"
sleep 20
after="$(max_restarts)"
printf 'max_restarts_before=%s max_restarts_after=%s\n' "${before:-0}" "${after:-0}"

if [ "${before:-0}" != "0" ] || [ "${after:-0}" != "0" ]; then
    echo "application containers have restarted" >&2
    exit 1
fi
