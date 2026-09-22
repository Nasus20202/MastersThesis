#!/bin/sh
set -eu

kubectl patch service app --type=merge \
    -p '{"spec":{"ports":[{"name":"http","port":80,"targetPort":8080,"protocol":"TCP"}]}}'
