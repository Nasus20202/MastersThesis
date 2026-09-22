#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec -n egress app -- timeout 6 wget -q -O - -T 3 http://dep/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        printf 'application reached its dependency\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the application to reach its dependency" >&2
exit 1
