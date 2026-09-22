#!/bin/sh
set -eu

if ! kubectl get deployment -n egress other >/dev/null 2>&1; then
    echo "the other workload deployment is missing from namespace egress" >&2
    exit 1
fi

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))
dependency_ok=false

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec -n egress app -- timeout 6 wget -q -O - -T 3 http://dep/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        dependency_ok=true
        break
    fi
    sleep 2
done

if [ "$dependency_ok" != true ]; then
    echo "the application cannot reach its dependency" >&2
    exit 1
fi

attempt=0
while [ "$attempt" -lt 3 ]; do
    if kubectl exec -n egress app -- timeout 6 wget -q -O - -T 3 http://other/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        echo "the application can still reach a disallowed workload, so egress is not restricted" >&2
        exit 1
    fi
    attempt=$((attempt + 1))
done

printf 'dependency is reachable and disallowed egress remains blocked\n'
