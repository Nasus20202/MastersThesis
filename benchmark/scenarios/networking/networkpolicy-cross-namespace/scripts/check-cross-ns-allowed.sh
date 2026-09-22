#!/bin/sh
set -eu

url='http://server.server-ns.svc.cluster.local/'
timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec -n client-ns allowed-client -- timeout 6 wget -q -O - -T 3 "$url" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        printf 'intended client reached the server in the other namespace\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the intended client to reach the server across namespaces" >&2
exit 1
