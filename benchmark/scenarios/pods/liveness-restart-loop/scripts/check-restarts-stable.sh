#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

ready=''
while [ "$(date +%s)" -lt "$deadline" ]; do
    ready="$(kubectl get deployment app -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    if [ -n "$ready" ] && [ "$ready" != "0" ]; then
        break
    fi
    sleep 2
done

if [ -z "$ready" ] || [ "$ready" = "0" ]; then
    echo "application Pods never became Ready" >&2
    exit 1
fi

markers="$(container_markers app=app)"
if printf '%s\n' "$markers" | grep -qE 'CrashLoopBackOff|Error'; then
    echo "an application container is still crashing" >&2
    exit 1
fi

before="$(restart_counts app=app)"
if [ -z "$before" ]; then
    echo "no application Pods found" >&2
    exit 1
fi
if printf '%s' "$before" | grep -qE '[1-9]'; then
    echo "an application container has already restarted" >&2
    exit 1
fi

sleep 15
after="$(restart_counts app=app)"
printf 'restarts_before=%s restarts_after=%s\n' "$(printf '%s' "$before" | tr '\n' ',')" "$(printf '%s' "$after" | tr '\n' ',')"
test "$before" = "$after"
