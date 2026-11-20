"""Training on Modal: one container and GPU per variant, all in parallel.

The local entrypoint refuses uncommitted changes, renders the attempts of a
manifest (``finetune.runs.prepare``) and uploads them to the results volume
under ``<run>/``. Each container gets this checkout's ``finetune`` package and
``config.yaml``, and trains one variant (``finetune.training.run``) into
``<run>/<variant>/``. GPU, memory, timeout and versions come from
``config.yaml``.

Every run costs money: start one only with the researcher's approval
(AGENTS.md). From ``finetune/``:

    uv run modal run -m finetune.modal --manifest <manifest.yaml> \
        --variants qwen35-4b-lora,gemma-4-e4b-lora --gpu H100
    uv run modal volume get finetune-results <run> <dir>
"""

import datetime
import logging
import sys
import tempfile
from pathlib import Path

import modal

from finetune import config, logs
from finetune.data import manifest as manifests
from finetune.runs import prepare
from finetune.runs.models import VARIANTS, Variant

SETTINGS = config.CONFIG.modal
log = logging.getLogger(__name__)
REMOTE_CONFIG = "/root/config.yaml"
RESULTS = Path("/results")

image = (
    # The local interpreter's version, which uv takes from .python-version.
    modal.Image.debian_slim(
        python_version=f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"
    )
    # Triton (flash-linear-attention) compiles its launchers with gcc.
    .apt_install("build-essential")
    .uv_sync(str(config.PATH.parent), extras=["peft"], uv_version=SETTINGS.uv)
    .env({"HF_HOME": "/hf", "FINETUNE_CONFIG": REMOTE_CONFIG})
    .add_local_file(config.PATH, REMOTE_CONFIG)
    .add_local_python_source("finetune")
)
app = modal.App(SETTINGS.app, image=image)
results = modal.Volume.from_name(SETTINGS.results_volume, create_if_missing=True)
hf_cache = modal.Volume.from_name(SETTINGS.cache_volume, create_if_missing=True)


@app.function(
    gpu=SETTINGS.gpu,
    cpu=SETTINGS.cpu,
    memory=SETTINGS.memory_mib,
    timeout=SETTINGS.timeout_minutes * 60,
    volumes={str(RESULTS): results, "/hf": hf_cache},
)
def train_variant(run_id: str, variant: str, steps: int) -> None:
    from finetune.training import run

    logs.setup()
    out = RESULTS / run_id
    try:
        run.run(prepare.job(Variant.parse(variant), out / "data", out / variant, steps))
    finally:
        results.commit()  # the run's log and metrics, also after a failure


@app.local_entrypoint()
def main(
    manifest: str,
    variants: str = ",".join(v.name for v in VARIANTS),
    gpu: str = SETTINGS.gpu,
    steps: int = config.CONFIG.training.steps,
) -> None:
    names = variants.split(",")
    parsed = [Variant.parse(name) for name in names]
    if prepare.provenance().uncommitted_changes:
        raise SystemExit("uncommitted changes; commit them so the run is traceable")
    run_id = datetime.datetime.now(datetime.UTC).strftime("run-%Y%m%d-%H%M%S")
    logs.setup()
    with tempfile.TemporaryDirectory() as tmp:
        session = prepare.prepare(Path(tmp), manifests.read(Path(manifest)), parsed, steps, gpu)
        with results.batch_upload() as batch:
            batch.put_directory(tmp, f"/{run_id}")
    log.info(
        "%s: %s at %s on %s",
        run_id,
        ", ".join(names),
        session.provenance.commit[:12],
        gpu,
    )
    outcomes = train_variant.with_options(gpu=gpu).starmap(
        ((run_id, name, steps) for name in names), return_exceptions=True
    )
    for name, outcome in zip(names, outcomes, strict=True):
        if outcome is None:
            log.info("%s: ok", name)
        else:
            log.error("%s: %s", name, outcome)
    log.info("download: uv run modal volume get %s %s <dir>", SETTINGS.results_volume, run_id)
