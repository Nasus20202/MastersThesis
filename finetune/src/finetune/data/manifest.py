"""Manifests: one file per training run (YAML, paths relative to the
repository), so every run's inputs are kept in the repository.

- ``rows``: a frozen dataset to train on, or ``train``: attempt files;
- ``eval``: attempt files, or a folder whose ``manifest.yaml`` lists them;
- ``variants``: what to train (empty: every variant).

Every attempt must be clean (D-054), and training attempts on validation
scenarios are refused (D-055)."""

from __future__ import annotations

from pathlib import Path

import yaml

from finetune.config import CONFIG, ROOT
from finetune.data import trajectories
from finetune.data.models import Manifest


def read(path: Path) -> Manifest:
    raw = yaml.safe_load(path.read_text())
    evals = raw.get("eval") or []
    if isinstance(evals, str):  # a folder that lists its attempts
        evals = yaml.safe_load((ROOT / evals / "manifest.yaml").read_text())["eval"]
    rows = ROOT / raw["rows"] if raw.get("rows") else None
    if rows is not None and not (rows / "freeze.json").is_file():
        raise ValueError(f"{rows} is not a frozen dataset")
    if rows is not None and raw.get("train"):
        raise ValueError("a manifest trains on rows or on train attempts, not both")
    manifest = Manifest(
        path.resolve(),
        [ROOT / p for p in raw.get("train") or []],
        [ROOT / p for p in evals],
        rows,
        list(raw.get("variants") or []),
    )
    unclean = [p for p in manifest.train + manifest.eval if not trajectories.is_clean(p)]
    if unclean:
        raise ValueError(f"attempts without full score or ended by the harness: {unclean}")
    leaked = [
        p for p in manifest.train if trajectories.load(p).scenario_id in CONFIG.validation_scenarios
    ]
    if leaked:
        raise ValueError(f"training attempts on validation scenarios: {leaked}")
    return manifest
