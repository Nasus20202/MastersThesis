#!/usr/bin/env bash

set -euo pipefail

CLUSTERS=$(kind get clusters)
while IFS= read -r CLUSTER; do
  case "$CLUSTER" in
    benchmark*) kind delete cluster --name "$CLUSTER" ;;
  esac
done <<< "$CLUSTERS"

CONTAINERS=$(docker ps -aq --filter name=sandbox)
while IFS= read -r CONTAINER; do
  if [[ -n "$CONTAINER" ]]; then
    docker rm --force "$CONTAINER"
  fi
done <<< "$CONTAINERS"

NETWORKS=$(docker network ls -q --filter name=sandbox-network)
while IFS= read -r NETWORK; do
  if [[ -n "$NETWORK" ]]; then
    docker network rm "$NETWORK"
  fi
done <<< "$NETWORKS"
