#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec -n egress app -- timeout 8 nslookup dep 2>/dev/null | grep -Eq 'Name:.*dep\.egress\.svc\.cluster\.local'; then
        printf 'application resolved dep.egress.svc.cluster.local via cluster DNS\n'
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the application to resolve its dependency through DNS" >&2
exit 1
