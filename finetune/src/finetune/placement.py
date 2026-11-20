"""Where a model's weights live when it trains on several GPUs.

The decoder layers are cut into contiguous blocks, one per GPU, so that every
GPU holds about the same bytes: the weights as loaded for training plus the
layer input that gradient checkpointing keeps at the longest sequence.
Everything that is not a decoder layer (embeddings, the output head, Gemma's
unused vision and audio towers) stays on GPU 0. The GPUs run one after the
other, so the split adds memory, not speed.

Gemma 4's per-layer embedding table (2.8B of E4B's 7.9B parameters) is frozen
and only looked up, so it is kept in host memory and looked up on the CPU.
"""

from __future__ import annotations

import torch

# Left in 16-bit by the 4-bit quantization: the output head and the towers.
NOT_QUANTIZED = ["lm_head", "vision", "audio", "visual"]
# Gemma's per-layer embeddings: frozen and only looked up.
HOST_TABLES = ("embed_tokens_per_layer",)


def decoder_layers(model) -> str:
    """Name of the language model's decoder layers: the largest module list."""
    lists = [
        (sum(p.numel() for p in m.parameters()), name)
        for name, m in model.named_modules()
        if isinstance(m, torch.nn.ModuleList)
    ]
    return max(lists)[1]


def placement_units(model, layers: str) -> list[str]:
    """Submodules that together hold every weight once: each decoder layer and
    every branch of the module tree beside the path down to them."""
    units = []
    path = layers.split(".")
    for depth in range(len(path)):
        parent = ".".join(path[:depth])
        module = model.get_submodule(parent)
        if next(module.parameters(recurse=False), None) is not None:
            raise ValueError(f"{parent or 'the model'} holds weights outside its submodules")
        for child, _ in module.named_children():
            name = f"{parent}.{child}" if parent else child
            if name != ".".join(path[: depth + 1]):
                units.append(name)
    units += [f"{layers}.{i}" for i in range(len(model.get_submodule(layers)))]
    return units


def weight_bytes(model, unit: str, precision: str, seen: set[int] | None = None) -> float:
    """Bytes of a unit's weights on the GPU once loaded and prepared for training.

    Host tables count zero, and weights already in ``seen`` (tied to another
    unit) too. With ``lora`` every weight takes two bytes (float16). With
    ``qlora`` the decoder projections take half a byte per weight, the other
    matrices two bytes and the 1-D weights (norms, biases) four (float32;
    ``train.frozen_matrices_to``).
    """
    seen = set() if seen is None else seen
    total = 0.0
    for name, module in model.get_submodule(unit).named_modules(prefix=unit):
        if name.endswith(HOST_TABLES):
            continue
        for param_name, p in module.named_parameters(recurse=False):
            if id(p) in seen:
                continue
            seen.add(id(p))
            if precision == "lora":
                size = 2
            elif (
                isinstance(module, torch.nn.Linear)
                and param_name == "weight"
                and not any(s in name for s in NOT_QUANTIZED)
            ):
                size = 0.5
            else:
                size = 2 if p.dim() >= 2 else 4
            total += p.numel() * size
    return total


def device_map(model, precision: str, gpus: int, tokens: int = 0) -> dict[str, int]:
    """GPU of every placement unit: decoder layers in contiguous blocks so that
    every GPU holds about the same bytes, everything else on GPU 0.

    A decoder layer counts its weights plus the input that gradient
    checkpointing keeps for it at ``tokens`` tokens (float16 hidden states).
    """
    layers = decoder_layers(model)
    units = placement_units(model, layers)
    seen: set[int] = set()
    sizes = {u: weight_bytes(model, u, precision, seen) for u in units}
    blocks = [u for u in units if u.startswith(layers + ".")]
    kept_input = tokens * model.config.get_text_config().hidden_size * 2
    for block in blocks:
        sizes[block] += kept_input
    placed = {u: 0 for u in units}
    used = sum(sizes[u] for u in units if u not in blocks)
    share = sum(sizes.values()) / gpus
    gpu = 0
    for block in blocks:
        if gpu < gpus - 1 and used + sizes[block] / 2 > share * (gpu + 1):
            gpu += 1
        placed[block] = gpu
        used += sizes[block]
    return placed


def summary(placement: dict[str, int]) -> dict[str, list[str]]:
    """Units per GPU, with consecutive decoder layers written as one range."""
    out: dict[str, list[str]] = {}
    for gpu in sorted(set(placement.values())):
        units = [u for u, g in placement.items() if g == gpu]
        layers = [u for u in units if u.rsplit(".", 1)[-1].isdigit()]
        listed = [u for u in units if u not in layers]
        if layers:
            prefix, first = layers[0].rsplit(".", 1)
            listed.append(f"{prefix}.{first}-{layers[-1].rsplit('.', 1)[1]}")
        out[str(gpu)] = listed
    return out


class HostLookup(torch.nn.Module):
    """A frozen embedding table looked up on the CPU; returns on the ids' device.

    The table is held in a list, so it is not a registered submodule: moving
    the model, the 4-bit preparation, parameter counts and saving all leave it
    alone.
    """

    def __init__(self, table: torch.nn.Module):
        super().__init__()
        self.table = [table.cpu().requires_grad_(False)]

    def forward(self, ids: torch.Tensor) -> torch.Tensor:
        return self.table[0](ids.cpu()).to(ids.device)


def tables_to_host(model) -> list[str]:
    """Replaces the host tables of ``model`` by ``HostLookup``; their names."""
    from accelerate.hooks import remove_hook_from_module

    names = [n for n, _ in model.named_modules() if n.endswith(HOST_TABLES)]
    for name in names:
        parent, _, attr = name.rpartition(".")
        # The hook from loading would move the ids back to the table's GPU.
        table = remove_hook_from_module(model.get_submodule(name))
        setattr(model.get_submodule(parent), attr, HostLookup(table))
    if names and torch.cuda.is_available():
        torch.cuda.empty_cache()
    return names
