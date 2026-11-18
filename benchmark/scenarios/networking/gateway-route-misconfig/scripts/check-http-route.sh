#!/bin/sh
set -eu

host="${1:?host required}"
path="${2:?path required}"
probe=route-probe

cleanup() {
    # Remove the probe and its events so the agent never sees a grader Pod.
    kubectl delete pod "$probe" --ignore-not-found --now --timeout=60s >/dev/null 2>&1 || true
    kubectl delete events --field-selector "involvedObject.kind=Pod,involvedObject.name=$probe" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

kubectl delete pod "$probe" --ignore-not-found >/dev/null 2>&1 || true
kubectl run "$probe" --image=busybox:1.37.0 --restart=Never --command -- sleep 3600 >/dev/null
kubectl wait --for=condition=Ready "pod/$probe" --timeout=120s >/dev/null

timeout_seconds=120
deadline=$(( $(date +%s) + timeout_seconds ))
# The route must keep working for several probes in a row: right after a
# change the gateway can still serve the previous route for a few seconds.
required_successes=8
successes=0

while [ "$(date +%s)" -lt "$deadline" ]; do
    if kubectl exec "$probe" -- wget -q -O - --header="Host: $host" "http://traefik.traefik.svc.cluster.local$path" 2>/dev/null | grep -q 'Welcome to nginx!'; then
        successes=$((successes + 1))
        if [ "$successes" -ge "$required_successes" ]; then
            printf 'request to %s%s reached the application %s times in a row\n' "$host" "$path" "$successes"
            exit 0
        fi
    else
        successes=0
    fi
    sleep 2
done

echo "timed out waiting for host $host path $path to reach the application" >&2
exit 1
