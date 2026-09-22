#!/bin/sh
set -eu

host="${1:?host required}"
path="${2:?path required}"
probe=route-probe

cleanup() {
    kubectl delete pod "$probe" --ignore-not-found --wait=false >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

kubectl delete pod "$probe" --ignore-not-found >/dev/null 2>&1 || true
kubectl run "$probe" --image=busybox:1.37.0 --restart=Never --command -- sleep 3600 >/dev/null
kubectl wait --for=condition=Ready "pod/$probe" --timeout=120s >/dev/null

timeout_seconds=90
deadline=$(( $(date +%s) + timeout_seconds ))

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec "$probe" -- wget -q -O - --header="Host: $host" "http://traefik.traefik.svc.cluster.local$path" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        printf 'request to %s%s reached the application\n' "$host" "$path"
        exit 0
    fi
    sleep 2
done

echo "timed out waiting for host $host path $path to reach the application" >&2
exit 1
