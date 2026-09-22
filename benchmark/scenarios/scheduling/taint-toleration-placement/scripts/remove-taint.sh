#!/bin/sh
set -eu

node="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
kubectl taint node "$node" dedicated=gpu:NoSchedule-
