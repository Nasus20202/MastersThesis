#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","env":[{"name":"APP_PASSWORD","valueFrom":{"secretKeyRef":{"name":"app-secret","key":"missing-key"}}}]}]}}}}'
