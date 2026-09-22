#!/bin/sh
set -eu

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    pod="$(kubectl get pods -l app=app -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
    if [ -n "$pod" ]; then
        if kubectl exec "$pod" -- sh -c 'test -s /var/run/secrets/kubernetes.io/serviceaccount/token && wget -q -O /dev/null --no-check-certificate --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config' >/dev/null 2>&1; then
            printf 'service account read configmap/app-config through the API\n'
            exit 0
        fi
    fi
    sleep 2
done

echo "timed out waiting for the in-cluster API call to succeed" >&2
exit 1
