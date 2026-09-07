# Benchmark

The benchmark consists of reproducible technical incidents in a controlled environment. A scenario starts from a known-good state, injects a real fault and verifies whether the model restores the required behaviour.

A scenario is accepted only when the clean state passes, the injected fault fails and an approved repair passes again.

The model sees the task and tools for the current condition. It must not see the fault definition, expected root cause, verifier logic or source references.

Documentation-dependent scenarios must trace to the frozen source corpus. RAG searches the complete approved corpus, not hand-picked scenario excerpts.

The final set should cover varied technical areas and include general, documentation-dependent, multi-source and version-specific incidents.
