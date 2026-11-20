"""Kaggle kernel of the technical check (#24): a starter for ``finetune.check``.

Clones the repository at ``REF``, installs the locked environment and runs
``python -m finetune.check`` with ``ARGS`` (empty runs every variant; e.g.
``--variants gemma-4-e2b-qlora --steps 2 --sweep 4096`` is a smoke run).
Results go to /kaggle/working/results. Needs the "GPU T4 x2" accelerator and
internet.
"""

import subprocess

REF = "feat/24-finetune-technical-check"
ARGS = ""
REPOSITORY = "https://github.com/Nasus20202/MastersThesis"

SCRIPT = f"""
set -euo pipefail
git clone --quiet --depth 1 --branch {REF} {REPOSITORY} /tmp/repo
git -C /tmp/repo log -1 --format='%H %s'
pip install --quiet uv==0.13.9
cd /tmp/repo/finetune
# Outside /kaggle/working, so neither becomes kernel output.
export UV_PROJECT_ENVIRONMENT=/tmp/venv HF_HOME=/tmp/hf
uv sync --locked --extra peft
uv run --locked --no-sync python -m finetune.check --out /kaggle/working/results {ARGS}
"""

subprocess.run(["bash", "-c", SCRIPT], check=True)
