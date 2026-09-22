#!/bin/sh
set -eu

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))
expected_count="${1:-2}"

while [ "$(date +%s)" -lt "$deadline" ]; do
    addresses="$(kubectl get endpointslices -l kubernetes.io/service-name=app -o jsonpath='{range .items[*].endpoints[?(@.conditions.ready==true)]}{.addresses[0]}{"\n"}{end}' 2>/dev/null || true)"
    set -- $addresses
    if [ "$#" -eq "$expected_count" ]; then
        printf 'ready_endpoints=%s\n' "$#"
        exit 0
    fi
    sleep 1
done

echo "timed out waiting for exactly $expected_count ready Service endpoints" >&2
exit 1
