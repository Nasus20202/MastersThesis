#!/bin/sh
set -eu

url='http://server.server-ns.svc.cluster.local/'
timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    allowed=false
    blocked=false
    if kubectl exec -n client-ns allowed-client -- timeout 6 wget -q -O - -T 3 "$url" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        allowed=true
    fi
    if kubectl exec -n other-ns blocked-client -- timeout 6 wget -q -O - -T 3 "$url" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        blocked=true
    fi
    if [ "$allowed" = true ] && [ "$blocked" = false ]; then
        printf 'intended client reaches the server and the other namespace is blocked\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the intended cross-namespace access state" >&2
exit 1
