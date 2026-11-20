"""Manifests: the attempt files a run trains and evaluates on (YAML, paths
relative to the repository). Every attempt must be clean (D-054), and
training attempts on validation scenarios are refused (D-055)."""

from __future__ import annotations

from pathlib import Path

import yaml

from finetune.config import CONFIG, ROOT
from finetune.data import trajectories
from finetune.data.models import Manifest


def read(path: Path) -> Manifest:
    raw = yaml.safe_load(path.read_text())
    manifest = Manifest(
        path.resolve(),
        [ROOT / p for p in raw.get("train") or []],
        [ROOT / p for p in raw.get("eval") or []],
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
