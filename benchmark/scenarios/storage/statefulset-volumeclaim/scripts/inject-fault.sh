#!/bin/sh
set -eu

kubectl delete statefulset app --ignore-not-found --wait=true
kubectl delete pod -l app=app --ignore-not-found --wait=true
kubectl delete pvc data-app-0 --ignore-not-found --wait=true
kubectl delete pvc data-app-1 --ignore-not-found --wait=true
kubectl create -f manifests/statefulset-bad.yaml
