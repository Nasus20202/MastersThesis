"""Records of the serving checks."""

from __future__ import annotations

import json
from dataclasses import asdict, dataclass


@dataclass(frozen=True)
class Conversion:
    ok: bool
    returncode: int
    log_tail: str


@dataclass(frozen=True)
class Perplexity:
    ok: bool
    ppl: float | None
    ppl_error: float | None
    lora_log: list[str]  # llama.cpp's own lines about loading the adapter
    log_tail: str


@dataclass(frozen=True)
class Served:
    """``served.json``: the conversion, then perplexity without and with the adapter."""

    convert: Conversion
    without_adapter: Perplexity | None = None
    with_adapter: Perplexity | None = None


@dataclass(frozen=True)
class Difference:
    """Where a served prompt and the HF render first differ, with context."""

    at: int
    served: str
    hf: str


@dataclass(frozen=True)
class TurnCheck:
    turn: int
    served_equals_hf_prompt: bool
    served_is_prefix_of_full: bool
    first_difference: Difference | None


@dataclass(frozen=True)
class AttemptCheck:
    student: str
    attempt: str
    checks: list[TurnCheck]

    @property
    def served_equals_hf_prompt(self) -> int:
        return sum(c.served_equals_hf_prompt for c in self.checks)

    @property
    def served_is_prefix_of_full(self) -> int:
        return sum(c.served_is_prefix_of_full for c in self.checks)

    def to_json(self) -> str:
        return json.dumps(
            {
                **asdict(self),
                "turns": len(self.checks),
                "served_equals_hf_prompt": self.served_equals_hf_prompt,
                "served_is_prefix_of_full": self.served_is_prefix_of_full,
            }
        )
