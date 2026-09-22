#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    allowed=false
    blocked=false
    if kubectl exec -n policy allowed-client -- timeout 6 wget -q -O - -T 3 http://server/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        allowed=true
    fi
    if kubectl exec -n policy blocked-client -- timeout 6 wget -q -O - -T 3 http://server/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        blocked=true
    fi
    if [ "$allowed" = true ] && [ "$blocked" = true ]; then
        printf 'both clients reach the server before any policy is applied\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for both clients to reach the server in the healthy state" >&2
exit 1
