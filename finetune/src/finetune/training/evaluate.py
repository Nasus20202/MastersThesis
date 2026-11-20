"""Target-token loss on held-out examples, with and without the adapter."""

from __future__ import annotations

import logging
from dataclasses import replace

import torch

from finetune.training.gpu import autocast, peak_gib
from finetune.training.lora import input_device
from finetune.training.loss import loss_and_count
from finetune.training.models import Evaluation, Tokens

log = logging.getLogger(__name__)


@torch.no_grad()
def evaluate(model, examples: list[Tokens]) -> Evaluation:
    model.eval()
    device = input_device(model)

    def mean_loss() -> float:
        total = count = 0.0
        for example in examples:
            with autocast():
                loss, n = loss_and_count(
                    model,
                    torch.tensor(example.ids, device=device),
                    torch.tensor(example.labels, device=device),
                )
            total += float(loss) * n
            count += n
        return total / max(count, 1)

    log.info("evaluating on %d examples", len(examples))
    result = Evaluation(len(examples), with_adapter=mean_loss(), without_adapter=0.0)
    with model.disable_adapter():
        result = replace(result, without_adapter=mean_loss())
    log.info(
        "evaluation loss %.4f with the adapter, %.4f without; peak %.1f GiB",
        result.with_adapter,
        result.without_adapter,
        peak_gib(),
    )
    return result
