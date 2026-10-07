#!/usr/bin/env bash

# Waits until the kis proxy at LLAMA_BASE_URL serves the MODEL_PROFILE model,
# then writes the session's stats and llama.cpp build to OUTPUT as provenance.
# Usage: kis-wait.sh OUTPUT.

set -euo pipefail

output="${1:?output file required}"
model=$(sed -n 's/^LLAMA_MODEL_NAME=//p' "$MODEL_PROFILE")

# The proxy answers 503 until a session is ready; a session serving another
# model is not used.
until stats=$(curl -sf "$LLAMA_BASE_URL/admin/stats") && [[ "$(jq -r .model <<<"$stats")" == "$model" ]]; do
  sleep 10
done

build=$(curl -sf "$LLAMA_BASE_URL/props" | jq -r .build_info)
mkdir -p "$(dirname "$output")"
jq -n --argjson stats "$stats" --arg build "$build" \
  '{image: env.KIS_IMAGE, build: $build, stats: $stats}' | tee "$output"
