#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    dep=false
    other=false
    resolved=false
    if kubectl exec -n egress app -- timeout 6 wget -q -O - -T 3 http://dep/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        dep=true
    fi
    if kubectl exec -n egress app -- timeout 6 wget -q -O - -T 3 http://other/ 2>/dev/null | grep -q 'Welcome to nginx!'; then
        other=true
    fi
    if kubectl exec -n egress app -- timeout 8 nslookup dep 2>/dev/null | grep -Eq 'Name:.*dep\.egress\.svc\.cluster\.local'; then
        resolved=true
    fi
    if [ "$dep" = true ] && [ "$other" = true ] && [ "$resolved" = true ]; then
        printf 'application reaches its dependency, resolves names and egress is open before any policy\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the healthy application egress state" >&2
exit 1
