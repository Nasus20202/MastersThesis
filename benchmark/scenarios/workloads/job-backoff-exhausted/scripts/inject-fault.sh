#!/bin/sh
set -eu

kubectl delete job app --ignore-not-found
kubectl create -f manifests/broken.yaml
