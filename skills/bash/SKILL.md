---
name: bash
description: Use Bash safely and readably for bounded inspection and changes inside the provided sandbox.
metadata:
  version: "1.0"
  scope: sandbox command execution
---

# Bash

Use Bash as the execution boundary for command-line work. Prefer commands that
make their inputs, outputs and failures visible.

- Keep commands short and bounded. Preserve stdout, stderr and the exit code as
  evidence; do not hide failures with `|| true`, broad redirection or discarded
  output.
- Quote variable expansions and data values. Use `--` where a command supports
  it, and avoid interpolating untrusted text into shell syntax.
- Separate inspection from mutation when possible. If a sequence depends on a
  successful check, stop or branch explicitly when that check fails.
- Prefer one narrow change over a broad replacement. Do not remove data or
  reset unrelated state to make a command succeed.
- Re-run a focused observation after a change. A successful command means only
  that the command ran; it does not prove the desired state.

Useful patterns include `command --help` for local syntax, `printf` for clear
diagnostic labels, and explicit conditionals for handling non-zero exits.
