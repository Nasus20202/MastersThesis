#!/usr/bin/env bash

set -euo pipefail

CLUSTER_PATTERN='^benchmark-.+-[a-f0-9]+$'
SANDBOX_PATTERN='^benchmark-.+-[a-f0-9]+-sandbox$'
NETWORK_PATTERN='^benchmark-.+-[a-f0-9]+-sandbox-network$'

while IFS= read -r CLUSTER; do
  if [[ "$CLUSTER" =~ $CLUSTER_PATTERN ]]; then
    kind delete cluster --name "$CLUSTER"
  fi
done < <(kind get clusters)

while IFS= read -r CONTAINER; do
  if [[ "$CONTAINER" =~ $SANDBOX_PATTERN ]]; then
    docker rm --force "$CONTAINER"
  fi
done < <(docker ps -a --format '{{.Names}}')

while IFS= read -r NETWORK; do
  if [[ "$NETWORK" =~ $NETWORK_PATTERN ]]; then
    docker network rm "$NETWORK"
  fi
done < <(docker network ls --format '{{.Name}}')
