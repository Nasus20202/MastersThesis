#!/bin/sh
set -eu

cluster_ip="$(kubectl get service app -o jsonpath='{.spec.clusterIP}' 2>/dev/null || true)"
selector="$(kubectl get service app -o jsonpath='{.spec.selector.app}' 2>/dev/null || true)"
port="$(kubectl get service app -o jsonpath='{.spec.ports[0].port}' 2>/dev/null || true)"
printf 'cluster_ip=%s selector=%s port=%s\n' "$cluster_ip" "$selector" "$port"
test "$cluster_ip" = "None"
test "$selector" = "app"
test "$port" = "80"

timeout_seconds=60
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    addresses="$(kubectl get endpointslices -l kubernetes.io/service-name=app -o jsonpath='{range .items[*].endpoints[?(@.conditions.ready==true)]}{.addresses[0]}{"\n"}{end}' 2>/dev/null || true)"
    set -- $addresses
    if [ "$#" -ge 2 ]; then
        printf 'ready_endpoints=%s\n' "$#"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for the headless Service to select the StatefulSet Pods" >&2
exit 1
