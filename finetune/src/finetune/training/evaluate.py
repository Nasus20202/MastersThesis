"""Target-token loss on held-out examples, with and without the adapter."""

from __future__ import annotations

import logging

import torch

from finetune.training.gpu import autocast, peak_gib
from finetune.training.lora import input_device
from finetune.training.loss import loss_and_count
from finetune.training.models import Evaluation, ExampleLoss, Tokens

log = logging.getLogger(__name__)


@torch.no_grad()
def evaluate(model, examples: list[Tokens]) -> Evaluation:
    model.eval()
    device = input_device(model)

    def losses() -> list[tuple[float, int]]:
        """Each example's mean loss and its number of target tokens."""
        found = []
        for example in examples:
            with autocast():
                loss, n = loss_and_count(
                    model,
                    torch.tensor(example.ids, device=device),
                    torch.tensor(example.labels, device=device),
                )
            found.append((float(loss), int(n)))
        return found

    def mean(found: list[tuple[float, int]]) -> float:
        return sum(loss * n for loss, n in found) / max(sum(n for _, n in found), 1)

    log.info("evaluating on %d examples", len(examples))
    adapted = losses()
    with model.disable_adapter():
        base = losses()
    result = Evaluation(
        len(examples),
        with_adapter=mean(adapted),
        without_adapter=mean(base),
        per_example=[
            ExampleLoss(e.source, n, a, b)
            for e, (a, n), (b, _) in zip(examples, adapted, base, strict=True)
        ],
    )
    log.info(
        "evaluation loss %.4f with the adapter, %.4f without; peak %.1f GiB",
        result.with_adapter,
        result.without_adapter,
        peak_gib(),
    )
    return result
