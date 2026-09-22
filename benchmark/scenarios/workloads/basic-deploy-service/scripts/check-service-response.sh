#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))
proxy_path='/api/v1/namespaces/basic-app/services/http:app:80/proxy/'

while [ "$(date +%s)" -lt "$deadline" ]; do
    body="$(kubectl get --raw "$proxy_path" 2>/dev/null || true)"
    if printf '%s' "$body" | grep -q 'Welcome to nginx!'; then
        printf '%s\n' 'service returned the expected nginx response'
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the Service to respond" >&2
exit 1
