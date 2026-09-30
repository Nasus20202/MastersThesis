// Package llamaenv creates llama.cpp clients from the LLAMA_* and EMBEDDING_*
// environment set by the model profile and retrieval.env.
package llamaenv

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
)

// Parallelism is the inference server capacity; the compose default is 1.
func Parallelism() (int, error) {
	value := strings.TrimSpace(os.Getenv(envLlamaParallel))
	if value == "" {
		value = "1"
	}
	parallelism, err := strconv.Atoi(value)
	if err != nil || parallelism < 1 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", envLlamaParallel, value)
	}
	return parallelism, nil
}

func NewChatClient() (inference.Client, error) {
	model := strings.TrimSpace(os.Getenv(envLlamaModelName))
	if model == "" {
		return nil, fmt.Errorf("%s is required", envLlamaModelName)
	}

	port := strings.TrimSpace(os.Getenv(envLlamaPort))
	if port == "" {
		port = "8080"
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", clientHost(), port),
		Model:   model,
		Metadata: inference.Metadata{
			Model:           model,
			Artifact:        modelArtifact(),
			Quantization:    strings.TrimSpace(os.Getenv(envLlamaModelQuant)),
			SHA256:          strings.TrimSpace(os.Getenv(envLlamaModelSHA256)),
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
	return inferenceClient, nil
}

func modelArtifact() string {
	repository := strings.TrimSpace(os.Getenv(envLlamaModelRepository))
	revision := strings.TrimSpace(os.Getenv(envLlamaModelRevision))
	file := strings.TrimSpace(os.Getenv(envLlamaModelFile))
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
	settings := make(map[string]string)
	for _, name := range llamaRuntimeEnvNames {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			settings[name] = value
		}
	}
	if len(settings) == 0 {
		return nil
	}
	return settings
}

func NewEmbeddingClient() (llama.Adapter, error) {
	model, port := os.Getenv(envEmbeddingModelName), os.Getenv(envEmbeddingPort)
	if model == "" || port == "" {
		return llama.Adapter{}, fmt.Errorf("%s and %s are required", envEmbeddingModelName, envEmbeddingPort)
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", clientHost(), port),
		Model:   model,
		Metadata: inference.Metadata{
			Provider:     "llama.cpp",
			Model:        model,
			Artifact:     fmt.Sprintf("%s@%s/%s", os.Getenv(envEmbeddingRepository), os.Getenv(envEmbeddingRevision), os.Getenv(envEmbeddingFile)),
			Quantization: os.Getenv(envEmbeddingQuant),
			SHA256:       os.Getenv(envEmbeddingSHA256),
		},
	})
	if err != nil {
		return llama.Adapter{}, fmt.Errorf("create embedding client: %w", err)
	}
	return llama.NewAdapter(client)
}

func clientHost() string {
	if host := strings.TrimSpace(os.Getenv(envLlamaClientHost)); host != "" {
		return host
	}
	return "127.0.0.1"
}
