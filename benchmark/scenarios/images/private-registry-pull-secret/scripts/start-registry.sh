#!/bin/sh
set -eu

scenario_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$scenario_dir"

user="${PRIVATE_REGISTRY_USER:-benchmark}"
password="${PRIVATE_REGISTRY_PASSWORD:-benchmark-pass}"
seed_image="${PRIVATE_REGISTRY_SEED_IMAGE:-docker.io/library/nginx:1.31.5}"
seed_ref="private/hello:1.0"

kubectl apply -f manifests/registry.yaml >/dev/null
kubectl -n registry rollout status deployment/registry --timeout=180s >/dev/null

port_log="$(mktemp)"
kubectl -n registry port-forward svc/registry 0:5000 >"$port_log" 2>&1 &
port_pid=$!

cleanup() {
    kill "$port_pid" 2>/dev/null || true
    rm -f "$port_log"
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
    --dest-creds "${user}:${password}" \
    "docker://${seed_image}" \
    "docker://localhost:${port}/${seed_ref}" >/dev/null
