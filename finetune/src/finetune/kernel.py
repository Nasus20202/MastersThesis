"""Image build step for the Modal image (``finetune.modal``): adds the
``cuda`` extra, causal-conv1d compiled from source, to the environment that
``uv_sync`` created.

It is a separate step only because ``uv_sync`` cannot request CPU and memory,
and the compile needs both. It runs in the build container, where neither
``config.yaml`` nor the rest of this package is set up, so it imports nothing
from ``finetune``.
"""

import subprocess

# Where Modal's uv_sync puts uv and the project (pyproject.toml, uv.lock).
UV = "/.uv/uv"
PROJECT = "/.uv"


def build() -> None:
    subprocess.run(
        [UV, "sync", f"--project={PROJECT}", "--no-install-workspace", "--frozen"]
        + ["--compile-bytecode", "--extra=peft", "--extra=cuda"],
        check=True,
    )
