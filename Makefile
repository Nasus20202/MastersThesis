GO ?= go
NPM ?= npm
BENCHMARK_DIR := benchmark
BENCHMARK_CONFIG := $(BENCHMARK_DIR)/config.env
MODEL_PROFILE ?= $(BENCHMARK_DIR)/model-profiles/gemma-4-e4b.env
SCENARIO ?= scenarios/
REPEAT ?= 1
BENCHMARK_PARALLEL ?= 4
VALIDATION_PARALLEL ?= 8

include $(BENCHMARK_CONFIG)

ifneq ($(strip $(MODEL_PROFILE)),)
include $(MODEL_PROFILE)
COMPOSE_ENV_FILES := --env-file $(BENCHMARK_CONFIG) --env-file $(MODEL_PROFILE)
else
COMPOSE_ENV_FILES := --env-file $(BENCHMARK_CONFIG)
endif

COMPOSE := docker compose $(COMPOSE_ENV_FILES) -f $(BENCHMARK_DIR)/docker-compose.yaml

export LLAMA_MODEL_REPOSITORY LLAMA_MODEL_REVISION LLAMA_MODEL_FILE LLAMA_MODEL_QUANTIZATION LLAMA_MODEL_SHA256 LLAMA_MODEL_DIR LLAMA_MODEL_NAME
export LLAMA_CONTEXT_SIZE LLAMA_GPU_LAYERS LLAMA_VULKAN_DEVICE LLAMA_PARALLEL
export LLAMA_FLASH_ATTN LLAMA_CACHE_TYPE_K LLAMA_CACHE_TYPE_V LLAMA_HOST LLAMA_PORT
export LLAMA_PUBLISH_HOST LLAMA_MODELS_MAX LLAMA_CLIENT_HOST
export LLAMA_REASONING LLAMA_REASONING_BUDGET

.DEFAULT_GOAL := help

.PHONY: help test format lint check download-models llama-start llama-stop llama-logs docker-cleanup build benchmark benchmark-validate

help:
	@printf '%s\n' 'Available commands:'
	@printf '  %-28s %s\n' \
		'make test' 'Run the benchmark Go tests.' \
		'make format' 'Format Go, Markdown, YAML, JSON, and other supported files.' \
		'make lint' 'Run Go vet, format checks, Prettier, and Renovate validation.' \
		'make check' 'Run test and lint checks.' \
		'make download-models' 'Download the configured model from Hugging Face.' \
		'make llama-start' 'Start the llama.cpp model router.' \
		'make llama-stop' 'Stop the llama.cpp model router.' \
		'make llama-logs' 'Follow llama.cpp model router logs.' \
		'make docker-cleanup' 'Remove benchmark Kind clusters and sandbox containers.' \
		'make build' 'Build the benchmark executable.' \
		'make benchmark' 'Run the default benchmark scenario.' \
		'make benchmark-validate' 'Validate benchmark scenarios with declared repairs.'
	@printf '%s\n' 'Benchmark parameters:'
	@printf '  %-28s %s\n' \
		'MODEL_PROFILE=PATH' 'Overlay a model profile, e.g. benchmark/model-profiles/qwen35-4b.env.' \
		'SCENARIO=PATH' 'Select scenario or validation directory/file (default: scenarios/).' \
		'REPEAT=N' 'Repeat each scenario or validation case (default: 1).' \
		'BENCHMARK_PARALLEL=N' 'Set agentic benchmark parallelism (default: 4).' \
		'VALIDATION_PARALLEL=N' 'Set validation parallelism (default: 8).'

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

docker-cleanup:
	./scripts/docker-cleanup.sh

build:
	cd $(BENCHMARK_DIR) && $(GO) build -o benchmark ./cmd/benchmark

benchmark:
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark --scenario $(SCENARIO) --parallel $(BENCHMARK_PARALLEL) --repeat $(REPEAT)

benchmark-validate:
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark --validate $(SCENARIO) --parallel $(VALIDATION_PARALLEL) --repeat $(REPEAT)

benchmark-go-test:
	cd $(BENCHMARK_DIR) && $(GO) test ./... -cover

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
