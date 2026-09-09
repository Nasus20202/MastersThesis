GO ?= go
NPM ?= npm
BENCHMARK_DIR := benchmark
BENCHMARK_CONFIG := $(BENCHMARK_DIR)/config.env
COMPOSE := docker compose --env-file $(BENCHMARK_CONFIG) -f $(BENCHMARK_DIR)/docker-compose.yaml

include $(BENCHMARK_CONFIG)

export LLAMA_MODEL_REPOSITORY LLAMA_MODEL_REVISION LLAMA_MODEL_FILE LLAMA_MODEL_DIR LLAMA_MODEL_NAME
export LLAMA_CONTEXT_SIZE LLAMA_GPU_LAYERS LLAMA_VULKAN_DEVICE LLAMA_PARALLEL
export LLAMA_FLASH_ATTN LLAMA_CACHE_TYPE_K LLAMA_CACHE_TYPE_V LLAMA_HOST LLAMA_PORT
export LLAMA_PUBLISH_HOST LLAMA_MODELS_MAX LLAMA_CLIENT_HOST

.DEFAULT_GOAL := help

.PHONY: help test format lint check download-models llama-start llama-stop llama-logs inference

help:
	@printf '%s\n' \
		'Available commands:' \
		'  make test            Run the benchmark Go tests.' \
		'  make format          Format Go, Markdown, YAML, JSON, and other supported files.' \
		'  make lint            Run Go vet, format checks, Prettier, and Renovate validation.' \
		'  make check           Run test and lint checks.' \
		'  make download-models Download the configured model from Hugging Face.' \
		'  make llama-start     Start the llama.cpp model router.' \
		'  make llama-stop      Stop the llama.cpp model router.' \
		'  make llama-logs      Follow llama.cpp model router logs.' \
		'  make inference       Run one inference through the Go client.'

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

llama-start:
	$(COMPOSE) up --detach

llama-stop:
	$(COMPOSE) down

llama-logs:
	$(COMPOSE) logs --follow llama-server

inference:
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/inference

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
