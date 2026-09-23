#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    pod="$(running_pod app=app)"
    if [ -n "$pod" ]; then
        uid="$(kubectl exec "$pod" -- id -u 2>/dev/null || true)"
        if [ -n "$uid" ] && [ "$uid" != "0" ]; then
            printf 'effective_uid=%s\n' "$uid"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for a non-root effective UID" >&2
exit 1
