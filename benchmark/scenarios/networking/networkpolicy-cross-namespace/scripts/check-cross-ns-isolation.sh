#!/bin/sh
set -eu

if ! kubectl get pod -n other-ns blocked-client >/dev/null 2>&1; then
    echo "blocked-client pod is missing from namespace other-ns" >&2
    exit 1
fi

url='http://server.server-ns.svc.cluster.local/'
timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))
allowed_ok=false

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec -n client-ns allowed-client -- timeout 6 wget -q -O - -T 3 "$url" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        allowed_ok=true
        break
    fi
    sleep 2
done

if [ "$allowed_ok" != true ]; then
    echo "the intended client cannot reach the server across namespaces" >&2
    exit 1
fi

attempt=0
while [ "$attempt" -lt 3 ]; do
    if kubectl exec -n other-ns blocked-client -- timeout 6 wget -q -O - -T 3 "$url" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        echo "a client in a non-intended namespace can reach the server" >&2
        exit 1
    fi
    attempt=$((attempt + 1))
done

printf 'intended namespace reaches the server and the other namespace remains blocked\n'
