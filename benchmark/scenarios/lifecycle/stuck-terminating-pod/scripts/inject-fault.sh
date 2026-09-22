#!/bin/sh
set -eu

kubectl patch pod app --type=merge \
    -p '{"metadata":{"finalizers":["example.com/hold"]}}'
kubectl delete pod app --wait=false
