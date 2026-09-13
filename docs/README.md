# Project documentation

This directory contains the research design, benchmark definition, technology choices, decisions and sources for the project.

## Documents

- [Research design](research-design.md) — research goal, questions, scope and comparison principles.
- [Benchmark](benchmark.md) — scenario structure, lifecycle, condition capabilities, visibility and verification.
- [Technology stack](tech-stack.md) — runner, model execution, Kubernetes environment and sandbox.
- [Decision log](decision-log.md) — canonical record of approved, provisional, open and superseded decisions.
- [Sources](sources.md) — technical documentation, papers and related work.
- [Glossary](glossary.md) — project terminology and links to further information.
- [Benchmark module](../benchmark/README.md) — local runner commands, logging and scenario configuration.

## Research evaluations

- [Kubernetes knowledge check](research/kubernetes-knowledge-check/README.md) — evaluated Gemma 4 E4B on Kubernetes troubleshooting probes without documentation context and with fixed excerpts from the pinned Kubernetes 1.37 documentation, preserving the prompts, criteria, raw responses and runtime metadata.
- [Baseline model comparison pilot](research/baseline-model-comparison/README.md) — Gemma 4 E2B, Gemma 4 E4B, Qwen3.5 4B and Qwen3.5 9B on 10 baseline `image-pull-failure` attempts each, with results and raw evidence.
- [Baseline incident evaluation](research/baseline-behaviour/README.md) — Gemma 4 E4B baseline on five development incidents, 10 attempts per incident, with merged raw evidence.
