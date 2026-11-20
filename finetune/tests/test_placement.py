"""GPU placement on CPU: which GPU every weight goes to, and host lookups."""

import pytest

torch = pytest.importorskip("torch")

from finetune import placement  # noqa: E402
from tests.conftest import tiny_gemma, tiny_qwen  # noqa: E402

VOCAB = 600


def test_host_lookup_returns_on_the_ids_device_and_survives_moves():
    table = torch.nn.Embedding(8, 4)
    lookup = placement.HostLookup(table)
    model = torch.nn.Sequential(lookup, torch.nn.Linear(4, 4))
    model.to("meta")
    assert lookup.table[0].weight.device.type == "cpu"
    assert not lookup.table[0].weight.requires_grad
    ids = torch.tensor([1, 2])
    assert torch.equal(lookup(ids), table(ids))


@pytest.mark.parametrize("build", [tiny_gemma, tiny_qwen])
def test_placement_covers_every_weight_once_and_keeps_layer_blocks(build):
    with torch.device("meta"):
        model = build(VOCAB)
    for gpus in (1, 2, 3):
        placed = placement.device_map(model, "lora", gpus, tokens=1024)
        for param, _ in model.named_parameters():
            owners = [u for u in placed if param.startswith(u + ".")]
            assert len(owners) == 1, param
        layers = placement.decoder_layers(model)
        blocks = [placed[f"{layers}.{i}"] for i in range(len(model.get_submodule(layers)))]
        assert blocks == sorted(blocks) and blocks[0] == 0
        assert all(g == 0 for u, g in placed.items() if not u.startswith(layers + "."))
        assert max(placed.values()) <= gpus - 1


def test_placement_balances_bytes_and_counts_tied_weights_once():
    class Tied(torch.nn.Module):
        def __init__(self):
            super().__init__()
            self.embed = torch.nn.Embedding(1000, 8)
            self.layers = torch.nn.ModuleList(torch.nn.Linear(8, 8) for _ in range(10))
            self.head = torch.nn.Linear(8, 1000, bias=False)
            self.head.weight = self.embed.weight
            self.config = type("C", (), {"get_text_config": lambda self: self, "hidden_size": 8})()

    model = Tied()
    seen: set[int] = set()
    assert placement.weight_bytes(model, "embed", "lora", seen) == 1000 * 8 * 2
    assert placement.weight_bytes(model, "head", "lora", seen) == 0
    # The embedding (16000 bytes) outweighs the ten layers (1440 bytes): one
    # GPU for it, the layers on the other.
    placed = placement.device_map(model, "lora", 2)
    assert {placed[f"layers.{i}"] for i in range(10)} == {1}
    # With kept inputs of 2000 bytes per layer (2144 each), GPU 0 also takes
    # one: 18144 against 19296 bytes is the most even cut.
    placed = placement.device_map(model, "lora", 2, tokens=125)
    assert [placed[f"layers.{i}"] for i in range(10)] == [0] + [1] * 9


def test_qlora_counts_projections_at_half_a_byte_and_1d_weights_in_float32():
    model = torch.nn.Module()
    model.layers = torch.nn.ModuleList([torch.nn.Linear(4, 4)])
    model.lm_head = torch.nn.Linear(4, 4, bias=False)
    assert placement.weight_bytes(model, "layers.0", "qlora") == 16 * 0.5 + 4 * 4
    assert placement.weight_bytes(model, "lm_head", "qlora") == 16 * 2
    assert placement.weight_bytes(model, "layers.0", "lora") == 20 * 2


def test_placement_summary_writes_layer_ranges():
    placed = {"embed": 0, "layers.0": 0, "layers.1": 0, "layers.2": 1, "layers.3": 1}
    assert placement.summary(placed) == {"0": ["embed", "layers.0-1"], "1": ["layers.2-3"]}
