"""The learning-rate schedule: linear warm-up, then constant or cosine."""

import pytest

pytest.importorskip("torch")

from finetune.training.trainer import lr_scale  # noqa: E402


def scales(schedule: str, warmup: int = 2, total: int = 10) -> list[float]:
    return [lr_scale(step, warmup, total, schedule) for step in range(total)]


def test_warmup_is_linear_and_reaches_the_full_rate():
    assert scales("constant")[:3] == [0.5, 1.0, 1.0]
    assert scales("cosine")[:2] == [0.5, 1.0]


def test_constant_holds_the_rate_after_warmup():
    assert scales("constant")[2:] == [1.0] * 8


def test_cosine_decays_monotonically_to_zero_at_the_last_step():
    decay = scales("cosine")[1:]
    assert decay == sorted(decay, reverse=True)
    assert decay[-1] == pytest.approx(0.0, abs=1e-12)
    assert scales("cosine")[2 + 3] == pytest.approx(0.5, abs=0.1)  # about halfway


def test_cosine_stays_in_range_when_steps_exceed_the_total():
    assert lr_scale(50, 2, 10, "cosine") == pytest.approx(0.0, abs=1e-12)


def test_unknown_schedule_is_rejected():
    with pytest.raises(ValueError):
        lr_scale(5, 2, 10, "linear")
