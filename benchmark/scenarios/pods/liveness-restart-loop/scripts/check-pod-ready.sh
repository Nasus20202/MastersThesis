#!/bin/sh
set -eu

. "$(dirname "$0")/../../../common/lib.sh"

timeout_seconds=90
expected_replicas=2
deadline=$(( $(date +%s) + timeout_seconds ))
consecutive=0

while [ "$(date +%s)" -lt "$deadline" ]; do
    if workload_rollout_ready deployment app "$expected_replicas"; then
        consecutive=$((consecutive + 1))
        if [ "$consecutive" -ge 3 ]; then
            exit 0
        fi
    else
        consecutive=0
    fi
    sleep 2
done

echo "timed out waiting for $expected_replicas sustained Ready application replicas" >&2
exit 1
