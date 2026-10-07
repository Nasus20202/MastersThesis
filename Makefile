GO ?= go
NPM ?= npm
BENCHMARK_DIR := benchmark
LLAMA_CONFIG := $(BENCHMARK_DIR)/config.env
MODEL_PROFILE ?= $(BENCHMARK_DIR)/model-profiles/gemma-4-e4b.env
SCENARIO ?= scenarios/
VALIDATION ?= $(SCENARIO)
CORPUS ?= scenarios/
AGENT ?= all
REPEAT ?= 1
CONFIG ?=
TAG ?=
BENCHMARK_PARALLEL ?= 4
VALIDATION_PARALLEL ?= 4
RESUME ?=
RUNS ?=
BROWSER_RESULTS ?= results
BROWSER_SCENARIOS ?= scenarios
RETRIEVAL_CONFIG := $(BENCHMARK_DIR)/retrieval.env
PROBES ?= ../docs/research/kubernetes-knowledge-check/context-tasks.json
QUERIES ?= ../docs/research/retrieval-design/queries.json
CHUNKING ?= all
Q ?=
SEARCH_ARGS ?=
ANALYZE_ARGS ?=
LLAMA_DEVICE ?= vulkan

include $(LLAMA_CONFIG)
include $(RETRIEVAL_CONFIG)

ifneq ($(strip $(MODEL_PROFILE)),)
include $(MODEL_PROFILE)
COMPOSE_ENV_FILES := --env-file $(LLAMA_CONFIG) --env-file $(RETRIEVAL_CONFIG) --env-file $(MODEL_PROFILE)
else
COMPOSE_ENV_FILES := --env-file $(LLAMA_CONFIG) --env-file $(RETRIEVAL_CONFIG)
endif

COMPOSE_FILES := -f $(BENCHMARK_DIR)/docker-compose.yaml -f $(BENCHMARK_DIR)/docker-compose.$(LLAMA_DEVICE).yaml
COMPOSE := docker compose $(COMPOSE_ENV_FILES) $(COMPOSE_FILES)

export LLAMA_MODEL_REPOSITORY LLAMA_MODEL_REVISION LLAMA_MODEL_FILE LLAMA_MODEL_QUANTIZATION LLAMA_MODEL_SHA256 LLAMA_MODEL_DIR LLAMA_MODEL_NAME
export LLAMA_MODEL_DRAFT_REPOSITORY LLAMA_MODEL_DRAFT_REVISION LLAMA_MODEL_DRAFT_FILE LLAMA_MODEL_DRAFT_SHA256
export LLAMA_SPEC_TYPE LLAMA_DRAFT_DIR LLAMA_SPEC_DRAFT_N_MAX
export LLAMA_KV_UNIFIED_PER_SLOT LLAMA_GPU_LAYERS LLAMA_VULKAN_DEVICE LLAMA_PARALLEL
export LLAMA_FLASH_ATTN LLAMA_CACHE_TYPE_K LLAMA_CACHE_TYPE_V LLAMA_HOST LLAMA_PORT
export LLAMA_PUBLISH_HOST LLAMA_MODELS_MAX LLAMA_CLIENT_HOST
export LLAMA_REASONING LLAMA_REASONING_BUDGET LLAMA_BASE_URL LLAMA_DEVICE EMBEDDING_BASE_URL
export CORPUS_REPOSITORY CORPUS_REVISION CORPUS_SUBTREE CORPUS_DIR
export EMBEDDING_MODEL_REPOSITORY EMBEDDING_MODEL_REVISION EMBEDDING_MODEL_FILE EMBEDDING_MODEL_QUANTIZATION
export EMBEDDING_MODEL_SHA256 EMBEDDING_MODEL_NAME EMBEDDING_MODEL_DIR EMBEDDING_PORT EMBEDDING_CTX_SIZE

# FTS5 in github.com/mattn/go-sqlite3 is compiled in only with this tag.
export GOFLAGS := -tags=sqlite_fts5

BENCHMARK_CONFIG_ARGS := --config config.yaml
ifneq ($(strip $(CONFIG)),)
BENCHMARK_CONFIG_ARGS += --config $(abspath $(CONFIG))
endif
BENCHMARK_RESUME_ARGS :=
ifneq ($(strip $(RESUME)),)
BENCHMARK_RESUME_ARGS += --resume $(RESUME)
endif
BENCHMARK_SCENARIO_ARGS := $(foreach path,$(SCENARIO),--scenario $(path))
BENCHMARK_VALIDATION_ARGS := $(foreach path,$(VALIDATION),--validate $(path))
BENCHMARK_TAG_ARGS := $(foreach tag,$(TAG),--tag $(tag))
# LLAMA_BASE_URL and EMBEDDING_BASE_URL point at already running
# OpenAI-compatible llama-servers instead of starting the local ones.
BENCHMARK_LLAMA := $(if $(LLAMA_BASE_URL),,llama-start)
EMBEDDING_START := $(if $(EMBEDDING_BASE_URL),,embedding-start)
# The RAG agent embeds its search queries, so start the embedding service only
# when it is selected.
comma := ,
BENCHMARK_EMBEDDING := $(if $(filter all rag,$(subst $(comma), ,$(AGENT))),$(EMBEDDING_START))

.DEFAULT_GOAL := help

.PHONY: help test format format-check lint check download-models llama-start llama-stop llama-logs registry-start registry-stop kind-network docker-cleanup build browser benchmark benchmark-validate check-corpus benchmark-go-lint corpus-download embedding-start embedding-stop retrieval-index retrieval-queries retrieval-evaluate retrieval-search analyze-runs benchmark-list benchmark-merge

help:
	@printf '%s\n' 'Available commands:'
	@printf '  %-28s %s\n' \
		'make test' 'Run the benchmark Go tests.' \
		'make format' 'Format Go, Markdown, YAML, JSON, and other supported files.' \
		'make lint' 'Run Go vet, golangci-lint, and Renovate validation.' \
		'make check' 'Run test, lint, format and corpus checks.' \
		'make download-models' 'Download the configured model, drafter and embedding GGUF from Hugging Face.' \
		'make corpus-download' 'Fetch the frozen Kubernetes documentation corpus.' \
		'make embedding-start' 'Start the llama.cpp embedding server.' \
		'make embedding-stop' 'Stop the llama.cpp embedding server.' \
		'make retrieval-index' 'Build the retrieval indexes from the corpus.' \
		'make retrieval-queries' 'Generate the frozen retrieval evaluation queries.' \
		'make retrieval-evaluate' 'Evaluate the retrieval configurations on the query set.' \
		'make retrieval-search Q=TEXT' 'Search the corpus with the configured retrieval settings.' \
		'make analyze-runs' 'Score RAG searches in benchmark runs against scenario sources (ANALYZE_ARGS).' \
		'make llama-start' 'Start the llama.cpp model router.' \
		'make llama-stop' 'Stop the llama.cpp model router.' \
		'make llama-logs' 'Follow llama.cpp model router logs.' \
		'make registry-start' 'Start the pull-through image registry used by benchmark clusters.' \
		'make registry-stop' 'Stop the pull-through image registry.' \
		'make docker-cleanup' 'Remove benchmark Kind clusters, sandbox and setup containers.' \
		'make build' 'Build the benchmark, browser, retrieval and analyze executables.' \
		'make browser' 'Browse benchmark run history in a terminal UI.' \
		'make benchmark' 'Run the default benchmark scenario.' \
		'make benchmark-list' 'Print the scenario files selected by SCENARIO and TAG.' \
		'make benchmark-merge' 'Merge the RUNS of one configuration, e.g. shards or repeats, into a new run.' \
		'make benchmark-validate' 'Validate benchmark scenarios with declared repairs.' \
		'make check-corpus' 'Load every scenario and its validation cases.'
	@printf '%s\n' 'Benchmark parameters:'
	@printf '  %-28s %s\n' \
		'MODEL_PROFILE=PATH' 'Overlay a model profile, e.g. benchmark/model-profiles/qwen35-4b.env.' \
		'LLAMA_BASE_URL=URL' 'Use a running llama-server instead of starting one (default: unset).' \
		'EMBEDDING_BASE_URL=URL' 'Use a running embedding server instead of starting one (default: unset).' \
		'LLAMA_DEVICE=NAME' 'llama.cpp server device: vulkan, cuda or cpu (default: vulkan).' \
		'SCENARIO=PATH...' 'Select scenario directories/files; space-separated (default: scenarios/).' \
		'VALIDATION=PATH...' 'Select validation directories/files; space-separated (default: scenarios/).' \
		'AGENT=NAME[,NAME]' 'Select all, baseline, prompt, skill, or rag benchmark agents (default: all).' \
		'CONFIG=PATH' 'Overlay a benchmark YAML config file.' \
		'TAG=SELECTOR' 'Filter scenarios by tag selector, e.g. difficulty=hard (default: none).' \
		'REPEAT=N' 'Repeat each scenario or validation case (default: 1).' \
		'RUNS=RUN_ID...' 'Runs merged by benchmark-merge; space-separated.' \
		'RESUME=RUN_ID' 'Resume an incomplete benchmark run by ID.' \
		'BENCHMARK_PARALLEL=N' 'Set agentic benchmark parallelism (default: 4).' \
		'VALIDATION_PARALLEL=N' 'Set validation parallelism (default: 4).' \
		'BROWSER_RESULTS=PATH' 'Results directory for the browser (default: results).' \
		'BROWSER_SCENARIOS=PATH' 'Scenario corpus for browser task titles (default: scenarios).' \
		'CHUNKING=NAME' 'Index sections, windows or all (default: all).' \
		'PROBES=PATH' 'Knowledge-check probes for query generation (relative to benchmark/).' \
		'QUERIES=PATH' 'Query set for retrieval evaluation (relative to benchmark/).' \
		'SEARCH_ARGS=FLAGS' 'Override search settings, e.g. --mode lexical --k 3.' \
		'ANALYZE_ARGS=FLAGS' 'Run analysis flags, e.g. --condition RUN/rag --reference RUN/prompt --out DIR.'

test:
	$(MAKE) benchmark-go-test

format:
	$(MAKE) benchmark-go-format
	$(MAKE) prettier

format-check:
	$(MAKE) benchmark-go-format-check
	$(MAKE) prettier-check

lint:
	$(MAKE) benchmark-go-vet
	$(MAKE) benchmark-go-lint
	$(MAKE) renovate-check

check: test lint format-check check-corpus

download-models:
	./scripts/download-models.sh

llama-start: kind-network
	$(COMPOSE) up --detach --wait llama-server

llama-stop:
	$(COMPOSE) stop llama-server

llama-logs:
	$(COMPOSE) logs --follow llama-server

corpus-download:
	./scripts/download-corpus.sh

embedding-start:
	$(COMPOSE) up --detach --wait llama-embedding

embedding-stop:
	$(COMPOSE) stop llama-embedding

retrieval-index: $(EMBEDDING_START)
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/retrieval build-index $(BENCHMARK_CONFIG_ARGS) --chunking $(CHUNKING)

retrieval-queries: llama-start
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/retrieval generate-queries --probes $(PROBES) --out $(QUERIES)

retrieval-evaluate: $(EMBEDDING_START)
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/retrieval evaluate $(BENCHMARK_CONFIG_ARGS) --queries $(QUERIES)

retrieval-search: $(EMBEDDING_START)
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/retrieval search $(BENCHMARK_CONFIG_ARGS) $(SEARCH_ARGS) "$(Q)"

analyze-runs:
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/analyze $(ANALYZE_ARGS)

kind-network:
	@docker network create kind >/dev/null 2>&1 || true

registry-start: kind-network
	$(COMPOSE) up --detach docker-registry-cache quay-registry-cache ghcr-registry-cache

registry-stop:
	$(COMPOSE) stop docker-registry-cache quay-registry-cache ghcr-registry-cache

docker-cleanup:
	./scripts/docker-cleanup.sh

build:
	cd $(BENCHMARK_DIR) && $(GO) build -o benchmark ./cmd/benchmark
	cd $(BENCHMARK_DIR) && $(GO) build -o browser ./cmd/browser
	cd $(BENCHMARK_DIR) && $(GO) build -o retrieval ./cmd/retrieval
	cd $(BENCHMARK_DIR) && $(GO) build -o analyze ./cmd/analyze

browser:
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/browser --results $(BROWSER_RESULTS) --scenarios $(BROWSER_SCENARIOS)

benchmark: registry-start $(BENCHMARK_LLAMA) $(BENCHMARK_EMBEDDING)
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark $(BENCHMARK_CONFIG_ARGS) $(BENCHMARK_SCENARIO_ARGS) $(BENCHMARK_TAG_ARGS) --agent $(AGENT) --parallel $(BENCHMARK_PARALLEL) --repeat $(REPEAT) $(BENCHMARK_RESUME_ARGS)

benchmark-list:
	@cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark $(BENCHMARK_CONFIG_ARGS) $(BENCHMARK_SCENARIO_ARGS) $(BENCHMARK_TAG_ARGS) --list

benchmark-merge:
	@cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark $(foreach run,$(RUNS),--merge $(run))

benchmark-validate: registry-start
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark $(BENCHMARK_CONFIG_ARGS) $(BENCHMARK_VALIDATION_ARGS) $(BENCHMARK_TAG_ARGS) --parallel $(VALIDATION_PARALLEL) --repeat $(REPEAT)

check-corpus:
	cd $(BENCHMARK_DIR) && $(GO) run ./cmd/benchmark --check-corpus $(CORPUS)

benchmark-go-test:
	cd $(BENCHMARK_DIR) && $(GO) test ./... -cover

benchmark-go-vet:
	cd $(BENCHMARK_DIR) && $(GO) vet ./...

benchmark-go-lint:
	cd $(BENCHMARK_DIR) && $(GO) tool golangci-lint run ./...

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
