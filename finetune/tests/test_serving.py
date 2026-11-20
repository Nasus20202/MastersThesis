"""Router preset section and base-vs-adapted comparison, with requests mocked."""

from finetune import serving
from tests.conftest import ATTEMPT


def test_preset_section_points_the_adapted_name_at_the_same_file():
    section = serving.preset_section(
        "Qwen3.5-4B-Q4_K_M-adapted",
        "/models/Qwen3.5-4B-Q4_K_M.gguf",
        "/adapters/qwen.gguf",
        {"temp": "1.0", "top-k": "20"},
    )
    assert section.splitlines() == [
        "[Qwen3.5-4B-Q4_K_M-adapted]",
        "model = /models/Qwen3.5-4B-Q4_K_M.gguf",
        "lora = /adapters/qwen.gguf",
        "temp = 1.0",
        "top-k = 20",
    ]


def test_prompts_are_the_first_agent_turn():
    [item] = serving.prompts([ATTEMPT])
    assert [m["role"] for m in item["messages"]] == ["system", "user"]
    assert item["tools"][0]["function"]["name"] == "bash"


def test_compare_summarises_speed_and_draft_acceptance(monkeypatch):
    def complete(url, model, prompt):
        adapted = model.endswith("adapted")
        return {
            "text": "b" if adapted else "a",
            "tool_calls": None,
            "predicted_n": 10,
            "predicted_per_second": 20.0 if adapted else 30.0,
            "draft_n": 10,
            "draft_n_accepted": 5 if adapted else 8,
        }

    monkeypatch.setattr(serving, "complete", complete)
    result = serving.compare("http://x", "base", "adapted", serving.prompts([ATTEMPT, ATTEMPT]))
    assert result["base_summary"] == {
        "prompts": 2,
        "draft_acceptance": 0.8,
        "mean_tokens_per_second": 30.0,
    }
    assert result["adapted_summary"]["draft_acceptance"] == 0.5
    assert result["same_text"] == 0


def test_summary_without_drafts():
    assert (
        serving.summary(
            [{"draft_n": None, "draft_n_accepted": None, "predicted_per_second": None}]
        )["draft_acceptance"]
        is None
    )
