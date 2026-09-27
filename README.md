# Master's Thesis

Comparison of methods for adapting local language models to technical
troubleshooting tasks in a local environment, evaluated on reproducible
Kubernetes incidents.

The benchmark runs each incident in a disposable local `kind` cluster. It
prepares a known-good environment, optionally injects a fault, gives the model
the task and its condition's capabilities, and then checks whether the system is
actually repaired. Conditions add different kinds of help to a common baseline:
a system prompt, reusable skills, retrieved documentation, fine-tuned weights or
a richer tool harness. Everything else, from the sandbox and runner to the
grading, stays the same, so the comparison isolates the adaptation method. Runs
are repeated because model behaviour is stochastic, and results are kept as they
come out, failures included. Models are served locally with `llama.cpp`.

## Running

```sh
make download-models   # fetch the pinned model and drafter
make benchmark         # run the scenarios
make browser           # browse run history
make check             # tests, lint, formatting and corpus validation
```

## Repository

- [`benchmark/`](benchmark/README.md): Go runner and terminal results browser.
- [`benchmark/scenarios/`](benchmark/scenarios): incident definitions, faults, expected repairs and grading.
- [`docs/`](docs/README.md): research design, benchmark contract, technology stack and decision log.
- [`docs/research/`](docs/research/): one study per topic, preserving its raw evidence.
