#!/bin/sh
set -eu

url='http://server.server-ns.svc.cluster.local/'
timeout_seconds=120
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if ! kubectl exec -n client-ns allowed-client -- timeout 6 wget -q -O - -T 3 "$url" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        printf 'the intended client can no longer reach the server across namespaces\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the cross-namespace policy to block the intended client" >&2
exit 1
