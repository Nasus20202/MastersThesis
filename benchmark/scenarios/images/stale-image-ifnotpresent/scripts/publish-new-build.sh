#!/bin/sh
set -eu

scenario_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$scenario_dir"

new_image="${MUTABLE_REGISTRY_NEW_IMAGE:-docker.io/library/nginx:1.29.0}"
new_ref="private/hello:stable"
digest_file="${KUBECONFIG}.published-digest"

kubectl -n registry rollout status deployment/registry --timeout=60s >/dev/null

port_log="$(mktemp)"
kubectl -n registry port-forward svc/registry 0:5000 >"$port_log" 2>&1 &
port_pid=$!

digest_tmp="$(mktemp)"
cleanup() {
    kill "$port_pid" 2>/dev/null || true
    rm -f "$port_log" "$digest_tmp"
}
trap cleanup EXIT

deadline=$(( $(date +%s) + 60 ))
port=""
while [ "$(date +%s)" -lt "$deadline" ]; do
    port="$(sed -n 's/.*127\.0\.0\.1:\([0-9]*\) ->.*/\1/p' "$port_log" | head -n1)"
    if [ -n "$port" ]; then
        break
    fi
    sleep 1
done
if [ -z "$port" ]; then
    echo 'could not establish a port-forward to the registry' >&2
    exit 1
fi

skopeo copy \
    --dest-tls-verify=false \
    --digestfile "$digest_tmp" \
    "docker://${new_image}" \
    "docker://localhost:${port}/${new_ref}" >/dev/null

digest="$(cat "$digest_tmp")"
if [ -z "$digest" ]; then
    echo 'could not read the published image digest' >&2
    exit 1
fi
printf '%s' "$digest" >"$digest_file"
