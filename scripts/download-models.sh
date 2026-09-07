#!/usr/bin/env bash

set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repository_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repository_root/benchmark"

hf_command=${HF:-hf}

: "${LLAMA_MODEL_REPOSITORY:?LLAMA_MODEL_REPOSITORY is required}"
: "${LLAMA_MODEL_REVISION:?LLAMA_MODEL_REVISION is required}"
: "${LLAMA_MODEL_FILE:?LLAMA_MODEL_FILE is required}"
: "${LLAMA_MODEL_DIR:?LLAMA_MODEL_DIR is required}"

if ! command -v "$hf_command" >/dev/null 2>&1; then
  printf '[download-models] error: %s was not found in PATH\n' "$hf_command" >&2
  exit 1
fi

mkdir -p "$LLAMA_MODEL_DIR"

printf '[download-models] repository: %s\n' "$LLAMA_MODEL_REPOSITORY"
printf '[download-models] revision: %s\n' "$LLAMA_MODEL_REVISION"
printf '[download-models] file: %s\n' "$LLAMA_MODEL_FILE"
printf '[download-models] destination: benchmark/%s\n' "$LLAMA_MODEL_DIR"

"$hf_command" download \
  "$LLAMA_MODEL_REPOSITORY" \
  "$LLAMA_MODEL_FILE" \
  --revision "$LLAMA_MODEL_REVISION" \
  --local-dir "$LLAMA_MODEL_DIR"

printf '[download-models] download complete\n'
