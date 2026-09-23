#!/usr/bin/env bash

# Downloads the configured model artifact and, when the profile declares one,
# the multi-token prediction drafter GGUF. Both files are checksum-verified.

set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repository_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repository_root/benchmark"

hf_command=${HF:-hf}

: "${LLAMA_MODEL_REPOSITORY:?LLAMA_MODEL_REPOSITORY is required}"
: "${LLAMA_MODEL_REVISION:?LLAMA_MODEL_REVISION is required}"
: "${LLAMA_MODEL_FILE:?LLAMA_MODEL_FILE is required}"
: "${LLAMA_MODEL_QUANTIZATION:?LLAMA_MODEL_QUANTIZATION is required}"
: "${LLAMA_MODEL_SHA256:?LLAMA_MODEL_SHA256 is required}"
: "${LLAMA_MODEL_DIR:?LLAMA_MODEL_DIR is required}"
: "${LLAMA_DRAFT_DIR:?LLAMA_DRAFT_DIR is required}"

if ! command -v "$hf_command" >/dev/null 2>&1; then
  printf '[download-models] error: %s was not found in PATH\n' "$hf_command" >&2
  exit 1
fi

# The drafter directory is bind-mounted by the server even when the selected
# profile has no drafter, so it is created regardless.
mkdir -p "$LLAMA_MODEL_DIR" "$LLAMA_DRAFT_DIR"

# A drafter is optional and selected by the model profile.
draft_repository=${LLAMA_MODEL_DRAFT_REPOSITORY:-}
if [[ -n "$draft_repository" ]]; then
  : "${LLAMA_MODEL_DRAFT_REVISION:?LLAMA_MODEL_DRAFT_REVISION is required when the profile declares a drafter}"
  : "${LLAMA_MODEL_DRAFT_FILE:?LLAMA_MODEL_DRAFT_FILE is required when the profile declares a drafter}"
  : "${LLAMA_MODEL_DRAFT_SHA256:?LLAMA_MODEL_DRAFT_SHA256 is required when the profile declares a drafter}"
fi

printf '[download-models] repository: %s\n' "$LLAMA_MODEL_REPOSITORY"
printf '[download-models] revision: %s\n' "$LLAMA_MODEL_REVISION"
printf '[download-models] file: %s\n' "$LLAMA_MODEL_FILE"
printf '[download-models] quantization: %s\n' "$LLAMA_MODEL_QUANTIZATION"
printf '[download-models] sha256: %s\n' "$LLAMA_MODEL_SHA256"
printf '[download-models] destination: benchmark/%s\n' "$LLAMA_MODEL_DIR"

"$hf_command" download \
  "$LLAMA_MODEL_REPOSITORY" \
  "$LLAMA_MODEL_FILE" \
  --revision "$LLAMA_MODEL_REVISION" \
  --local-dir "$LLAMA_MODEL_DIR"

actual_sha256=$(sha256sum "$LLAMA_MODEL_DIR/$LLAMA_MODEL_FILE" | awk '{print $1}')
if [[ "$actual_sha256" != "$LLAMA_MODEL_SHA256" ]]; then
  printf '[download-models] error: SHA-256 mismatch for %s (expected %s, got %s)\n' \
    "$LLAMA_MODEL_FILE" "$LLAMA_MODEL_SHA256" "$actual_sha256" >&2
  exit 1
fi

printf '[download-models] download complete; SHA-256 verified\n'

if [[ -n "$draft_repository" ]]; then
  printf '[download-models] draft repository: %s\n' "$draft_repository"
  printf '[download-models] draft revision: %s\n' "$LLAMA_MODEL_DRAFT_REVISION"
  printf '[download-models] draft file: %s\n' "$LLAMA_MODEL_DRAFT_FILE"
  printf '[download-models] draft destination: benchmark/%s\n' "$LLAMA_DRAFT_DIR"

  "$hf_command" download \
    "$draft_repository" \
    "$LLAMA_MODEL_DRAFT_FILE" \
    --revision "$LLAMA_MODEL_DRAFT_REVISION" \
    --local-dir "$LLAMA_DRAFT_DIR"

  # Downloading mirrors the repository layout; flatten the drafter so its path
  # does not depend on the upstream folder and the router preset stays simple.
  draft_name=$(basename -- "$LLAMA_MODEL_DRAFT_FILE")
  if [[ "$LLAMA_MODEL_DRAFT_FILE" != "$draft_name" ]]; then
    mv -f "$LLAMA_DRAFT_DIR/$LLAMA_MODEL_DRAFT_FILE" "$LLAMA_DRAFT_DIR/$draft_name"
    rmdir --ignore-fail-on-non-empty "$LLAMA_DRAFT_DIR/$(dirname -- "$LLAMA_MODEL_DRAFT_FILE")" 2>/dev/null || true
  fi

  actual_draft_sha256=$(sha256sum "$LLAMA_DRAFT_DIR/$draft_name" | awk '{print $1}')
  if [[ "$actual_draft_sha256" != "$LLAMA_MODEL_DRAFT_SHA256" ]]; then
    printf '[download-models] error: SHA-256 mismatch for %s (expected %s, got %s)\n' \
      "$draft_name" "$LLAMA_MODEL_DRAFT_SHA256" "$actual_draft_sha256" >&2
    exit 1
  fi

  printf '[download-models] draft download complete; SHA-256 verified\n'
fi
