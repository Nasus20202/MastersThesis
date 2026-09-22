#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"nodeSelector":{"disktype":"ssd"}}}}}'
