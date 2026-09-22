#!/bin/sh
set -eu

if ! kubectl get pod -n policy blocked-client >/dev/null 2>&1; then
    echo "blocked-client pod is missing from namespace policy" >&2
    exit 1
fi

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))
allowed_ok=false

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec -n policy allowed-client -- timeout 6 wget -q -O - -T 3 http://server/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        allowed_ok=true
        break
    fi
    sleep 2
done

if [ "$allowed_ok" != true ]; then
    echo "the intended client cannot reach the application server" >&2
    exit 1
fi

attempt=0
while [ "$attempt" -lt 3 ]; do
    if kubectl exec -n policy blocked-client -- timeout 6 wget -q -O - -T 3 http://server/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        echo "a non-intended client unexpectedly reached the application server" >&2
        exit 1
    fi
    attempt=$((attempt + 1))
done

printf 'allowed client reaches the server and the other client remains blocked\n'
