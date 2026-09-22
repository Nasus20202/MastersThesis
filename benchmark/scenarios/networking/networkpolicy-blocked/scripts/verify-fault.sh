#!/bin/sh
set -eu

timeout_seconds=120
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if ! kubectl exec -n policy allowed-client -- timeout 6 wget -q -O - -T 3 http://server/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        printf 'the previously working client is now blocked from the application server\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the NetworkPolicy to block the intended client" >&2
exit 1
