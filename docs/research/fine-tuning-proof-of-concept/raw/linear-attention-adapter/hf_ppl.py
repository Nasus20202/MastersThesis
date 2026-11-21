"""Perplexity of Qwen3.5 4B in transformers (CPU, float32) on the first
windows of a text, with and without a PEFT adapter, scored like
llama-perplexity: second half of each window."""
import json, math, sys
sys.modules["fla"] = None  # Triton-only kernels; use the PyTorch path on the CPU (as the tests do)
from pathlib import Path
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
from peft import PeftModel

text, adapter, ctx, windows = Path(sys.argv[1]), Path(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
name, rev = "Qwen/Qwen3.5-4B", "851bf6e806efd8d0a36b00ddf55e13ccb7b8cd0a"
tok = AutoTokenizer.from_pretrained(name, revision=rev)
ids = tok(text.read_text(), add_special_tokens=False).input_ids[: ctx * windows]
text.with_suffix(".scored.txt").write_text(tok.decode(ids))
model = AutoModelForCausalLM.from_pretrained(name, revision=rev, dtype=torch.float32).eval()

def ppl(m):
    nll, n = 0.0, 0
    with torch.no_grad():
        for w in range(windows):
            x = torch.tensor([ids[w * ctx : (w + 1) * ctx]])
            logits = m(input_ids=x).logits[0, :-1].float()
            lp = torch.log_softmax(logits, -1).gather(1, x[0, 1:, None])[:, 0]
            part = lp[ctx // 2 - 1 :]
            nll -= part.sum().item(); n += len(part)
    return math.exp(nll / n)

base = ppl(model)
print("without", base, flush=True)
with_adapter = ppl(PeftModel.from_pretrained(model, adapter).eval())
print("with", with_adapter, "ratio", with_adapter / base)
