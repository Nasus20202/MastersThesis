package agent

import (
	"fmt"
	"os"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/baseline"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
)

func NewBaselineFactory() (rootagent.Factory, error) {
	model := strings.TrimSpace(os.Getenv("LLAMA_MODEL_NAME"))
	if model == "" {
		return nil, fmt.Errorf("LLAMA_MODEL_NAME is required")
	}

	host := strings.TrimSpace(os.Getenv("LLAMA_CLIENT_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	port := strings.TrimSpace(os.Getenv("LLAMA_PORT"))
	if port == "" {
		port = "8080"
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", host, port),
		Model:   model,
		Metadata: inference.Metadata{
			Model:           model,
			Artifact:        modelArtifact(),
			Quantization:    strings.TrimSpace(os.Getenv("LLAMA_MODEL_QUANTIZATION")),
			SHA256:          strings.TrimSpace(os.Getenv("LLAMA_MODEL_SHA256")),
			RuntimeSettings: runtimeSettings(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create llama client: %w", err)
	}
	inferenceClient, err := llama.NewAdapter(client)
	if err != nil {
		return nil, fmt.Errorf("create inference adapter: %w", err)
	}
	loopConfig := common.DefaultConfig()
	return func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
		return baseline.New(inferenceClient, shell, loopConfig)
	}, nil
}

func modelArtifact() string {
	repository := strings.TrimSpace(os.Getenv("LLAMA_MODEL_REPOSITORY"))
	revision := strings.TrimSpace(os.Getenv("LLAMA_MODEL_REVISION"))
	file := strings.TrimSpace(os.Getenv("LLAMA_MODEL_FILE"))
	artifact := repository
	if revision != "" {
		artifact += "@" + revision
	}
	if file != "" {
		artifact += "/" + file
	}
	return artifact
}

func runtimeSettings() map[string]string {
	const names = "LLAMA_CONTEXT_SIZE LLAMA_GPU_LAYERS LLAMA_VULKAN_DEVICE LLAMA_PARALLEL LLAMA_FLASH_ATTN LLAMA_CACHE_TYPE_K LLAMA_CACHE_TYPE_V LLAMA_MODELS_MAX LLAMA_REASONING LLAMA_REASONING_BUDGET LLAMA_HOST LLAMA_PORT LLAMA_CLIENT_HOST"
	settings := make(map[string]string)
	for _, name := range strings.Fields(names) {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			settings[name] = value
		}
	}
	if len(settings) == 0 {
		return nil
	}
	return settings
}
