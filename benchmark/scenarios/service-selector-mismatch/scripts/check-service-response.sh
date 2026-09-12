#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))
proxy_path='/api/v1/namespaces/default/services/http:app:80/proxy/'

while [ "$(date +%s)" -lt "$deadline" ]; do
    body="$(kubectl get --raw "$proxy_path" 2>/dev/null || true)"
    if printf '%s' "$body" | grep -q 'Welcome to nginx!'; then
        printf '%s\n' 'service returned expected nginx response'
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the expected Service response" >&2
exit 1
