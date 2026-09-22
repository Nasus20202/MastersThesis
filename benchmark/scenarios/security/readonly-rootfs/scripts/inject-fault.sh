#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","securityContext":{"readOnlyRootFilesystem":true}}]}}}}'
