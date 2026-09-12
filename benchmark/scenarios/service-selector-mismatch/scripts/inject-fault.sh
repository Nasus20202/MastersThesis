#!/bin/sh
set -eu

kubectl patch service app --type=merge \
    -p '{"spec":{"selector":{"app":"does-not-match"}}}'
