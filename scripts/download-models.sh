#!/usr/bin/env bash

set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repository_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repository_root"

hf_command=${HF:-hf}
model_repository=${LLAMA_MODEL_REPOSITORY:-google/gemma-4-E4B-it-qat-q4_0-gguf}
model_revision=${LLAMA_MODEL_REVISION:-main}
model_file=${LLAMA_MODEL_FILE:-gemma-4-E4B_q4_0-it.gguf}
model_dir=${LLAMA_MODEL_DIR:-benchmark/models}

if ! command -v "$hf_command" >/dev/null 2>&1; then
  printf '[download-models] error: %s was not found in PATH\n' "$hf_command" >&2
  exit 1
fi

mkdir -p "$model_dir"

printf '[download-models] repository: %s\n' "$model_repository"
printf '[download-models] revision: %s\n' "$model_revision"
printf '[download-models] file: %s\n' "$model_file"
printf '[download-models] destination: %s\n' "$model_dir"

"$hf_command" download \
  "$model_repository" \
  "$model_file" \
  --revision "$model_revision" \
  --local-dir "$model_dir"

printf '[download-models] download complete\n'
