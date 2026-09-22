#!/bin/sh
set -eu

kubectl patch deployment app --type=strategic \
    -p '{"spec":{"template":{"spec":{"containers":[{"name":"client","env":[{"name":"PEER_HOST","value":"peer.prod"}]}]}}}}'
