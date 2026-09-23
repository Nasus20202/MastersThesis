#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=150
stable_seconds=20
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if workload_ready deployment app 1 >/dev/null; then
        sleep "$stable_seconds"
        if workload_ready deployment app 1 >/dev/null; then
            printf 'replicas=1 ready=1 for %ss\n' "$stable_seconds"
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for a stable Ready Pod" >&2
exit 1
