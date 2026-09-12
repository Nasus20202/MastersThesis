#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))
proxy_path='/api/v1/namespaces/default/services/http:app:80/proxy/'

while [ "$(date +%s)" -lt "$deadline" ]; do
    addresses="$(kubectl get endpointslices -l kubernetes.io/service-name=app -o jsonpath='{range .items[*].endpoints[?(@.conditions.ready==true)]}{.addresses[0]}{"\n"}{end}' 2>/dev/null || true)"
    body="$(kubectl get --raw "$proxy_path" 2>/dev/null || true)"
    if [ -z "$addresses" ] && ! printf '%s' "$body" | grep -q 'Welcome to nginx!'; then
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for the Service to lose its endpoints" >&2
exit 1
