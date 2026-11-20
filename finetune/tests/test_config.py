"""config.yaml refers to things that exist."""

from finetune.config import CONFIG, ROOT


def test_validation_scenarios_are_benchmark_scenarios():
    scenarios = {p.name for p in (ROOT / "benchmark" / "scenarios").glob("*/*") if p.is_dir()}
    assert CONFIG.validation_scenarios <= scenarios
