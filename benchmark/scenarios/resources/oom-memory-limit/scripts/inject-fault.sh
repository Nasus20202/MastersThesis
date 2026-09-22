#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","resources":{"limits":{"memory":"4Mi"}}}]}}}}'
