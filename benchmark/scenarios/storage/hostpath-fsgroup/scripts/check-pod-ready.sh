#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if workload_ready deployment app 3; then
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for three ready application Pods" >&2
exit 1
