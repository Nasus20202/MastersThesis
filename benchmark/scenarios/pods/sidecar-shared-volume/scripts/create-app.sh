#!/bin/sh
set -eu

deadline=$(( $(date +%s) + 60 ))
until kubectl get serviceaccount default >/dev/null 2>&1; do
    if [ "$(date +%s)" -ge "$deadline" ]; then
        echo "default ServiceAccount did not appear in time" >&2
        exit 1
    fi
    sleep 1
done

kubectl create -f manifests/app.yaml
