"""A random LoRA adapter on Qwen3.5 4B's Gated DeltaNet projections only, in
the key layout of the trained POC adapter. B is non-zero so the adapter
visibly changes the output if it is applied."""
import json, sys
from pathlib import Path
import torch
from safetensors.torch import save_file

out, scale = Path(sys.argv[1]), float(sys.argv[2])
trained = Path(sys.argv[3])  # trained adapter dir, for its config
cfg = json.loads((trained / "adapter_config.json").read_text())
MODULES = sys.argv[4].split(",")
cfg["target_modules"] = r"^(?!.*(?:vision|audio|visual)).*\.(?:" + "|".join(MODULES) + r")$"
hidden, key, value, r = 2560, 16 * 128, 32 * 128, cfg["r"]
ALL = {"in_proj_qkv": (hidden, 2 * key + value), "in_proj_z": (hidden, value), "out_proj": (value, hidden)}
shapes = {m: ALL[m] for m in MODULES}
linear = [i for i in range(32) if i % 4 != 3]  # every fourth layer is full attention
g = torch.Generator().manual_seed(0)
tensors = {}
for i in linear:
    for m, (d_in, d_out) in shapes.items():
        p = f"base_model.model.model.layers.{i}.linear_attn.{m}"
        tensors[f"{p}.lora_A.weight"] = (torch.rand(r, d_in, generator=g) * 2 - 1) / d_in**0.5
        tensors[f"{p}.lora_B.weight"] = torch.randn(d_out, r, generator=g) * scale
out.mkdir(parents=True, exist_ok=True)
(out / "adapter_config.json").write_text(json.dumps(cfg, indent=2))
save_file(tensors, out / "adapter_model.safetensors")
print(len(tensors), "tensors for", len(linear), "layers")
