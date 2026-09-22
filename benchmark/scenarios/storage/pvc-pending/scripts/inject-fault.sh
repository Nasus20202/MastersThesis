#!/bin/sh
set -eu

kubectl delete deployment app --ignore-not-found --wait=true
kubectl delete pod -l app=app --ignore-not-found --wait=true
kubectl delete pvc app-pvc --ignore-not-found --wait=true
kubectl create -f manifests/app-bad-pvc.yaml
