GO ?= go
NPM ?= npm
BENCHMARK_DIR := benchmark

.DEFAULT_GOAL := help

.PHONY: help test format lint check download-models

help:
	@printf '%s\n' \
		'Available commands:' \
		'  make test           Run the benchmark Go tests.' \
		'  make format         Format Go, Markdown, YAML, JSON, and other supported files.' \
		'  make lint           Run Go vet, format checks, Prettier, and Renovate validation.' \
		'  make check          Run test and lint checks.' \
		'  make download-models Download the configured model from Hugging Face.'

test:
	$(MAKE) benchmark-go-test

format:
	$(MAKE) benchmark-go-format
	$(MAKE) prettier

lint:
	$(MAKE) benchmark-go-vet
	$(MAKE) benchmark-go-format-check
	$(MAKE) prettier-check
	$(MAKE) renovate-check

check: test lint

download-models:
	./scripts/download-models.sh

benchmark-go-test:
	cd $(BENCHMARK_DIR) && $(GO) test ./...

benchmark-go-vet:
	cd $(BENCHMARK_DIR) && $(GO) vet ./...

benchmark-go-format:
	cd $(BENCHMARK_DIR) && $(GO) fmt ./...

benchmark-go-format-check:
	@test -z "$$(find $(BENCHMARK_DIR) -type f -name '*.go' -print0 | xargs -0 gofmt -l)"

prettier:
	$(NPM) exec --no -- prettier . --write --ignore-unknown

prettier-check:
	$(NPM) exec --no -- prettier . --check --ignore-unknown

renovate-check:
	$(NPM) exec --no -- renovate-config-validator --strict renovate.json
