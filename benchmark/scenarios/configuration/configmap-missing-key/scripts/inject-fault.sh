#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","env":[{"name":"CONFIG_MESSAGE","valueFrom":{"configMapKeyRef":{"name":"app-config","key":"missing"}}}]}]}}}}'
