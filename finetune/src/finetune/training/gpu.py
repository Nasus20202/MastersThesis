"""Precision and memory of the GPU (no-ops on the CPU, where tests run)."""

from __future__ import annotations

import contextlib
import gc

import torch


def compute_dtype() -> torch.dtype:
    """bfloat16 on the GPU, float32 on the CPU."""
    return torch.bfloat16 if torch.cuda.is_available() else torch.float32


def autocast():
    if torch.cuda.is_available():
        return torch.autocast("cuda", dtype=compute_dtype())
    return contextlib.nullcontext()


def sync() -> None:
    if torch.cuda.is_available():
        torch.cuda.synchronize()


def peak_gib() -> float:
    """Peak allocated GPU memory since the last reset."""
    return torch.cuda.max_memory_allocated() / 2**30 if torch.cuda.is_available() else 0.0


def allocated_gib() -> float:
    return torch.cuda.memory_allocated() / 2**30 if torch.cuda.is_available() else 0.0


def reset_peak() -> None:
    gc.collect()
    if torch.cuda.is_available():
        torch.cuda.empty_cache()
        torch.cuda.reset_peak_memory_stats()


def is_oom(error: BaseException) -> bool:
    return isinstance(error, torch.OutOfMemoryError) or "out of memory" in str(error).lower()


def oom_message(error: BaseException) -> str:
    return str(error).splitlines()[0][:300]
