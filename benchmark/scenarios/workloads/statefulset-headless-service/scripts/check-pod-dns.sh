#!/bin/sh
set -eu

probe=dns-probe

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
    resolved=true
    for ordinal in 0 1; do
        pod_ip="$(kubectl get pod "app-$ordinal" -o jsonpath='{.status.podIP}' 2>/dev/null || true)"
        if [ -z "$pod_ip" ]; then
            resolved=false
            continue
        fi
        output="$(kubectl exec "$probe" -- nslookup "app-$ordinal.app.default.svc.cluster.local" 2>/dev/null || true)"
        if ! printf '%s\n' "$output" | grep -q "$pod_ip"; then
            resolved=false
        fi
    done
    if [ "$resolved" = true ]; then
        printf 'per-Pod DNS resolved for app-0 and app-1\n'
        exit 0
    fi
    sleep 3
done

echo "timed out waiting for the StatefulSet per-Pod DNS names to resolve" >&2
exit 1
