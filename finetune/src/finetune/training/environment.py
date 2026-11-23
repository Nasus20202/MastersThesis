"""Software and GPU of the machine a run trained on."""

from __future__ import annotations

import contextlib
import platform
from importlib import metadata

import torch

from finetune.training.gpu import compute_dtype
from finetune.training.models import Environment, Gpu

PACKAGES = (
    "torch",
    "transformers",
    "peft",
    "bitsandbytes",
    "accelerate",
    "flash-linear-attention",
    "causal-conv1d",
    "triton",
)


def environment() -> Environment:
    packages = {}
    for name in PACKAGES:
        with contextlib.suppress(metadata.PackageNotFoundError):
            packages[name] = metadata.version(name)
    gpu = None
    if torch.cuda.is_available():
        props = torch.cuda.get_device_properties()
        gpu = Gpu(props.name, props.total_memory / 2**30)
    return Environment(
        python=platform.python_version(),
        packages=packages,
        dtype=str(compute_dtype()).removeprefix("torch."),
        gpu=gpu,
        cuda=torch.version.cuda if torch.cuda.is_available() else None,
    )
