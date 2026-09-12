#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","args":["nginx","-g","not_a_valid_directive on;"]}]}}}}'
