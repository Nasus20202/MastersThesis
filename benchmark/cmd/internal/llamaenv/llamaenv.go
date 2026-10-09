// Package llamaenv creates inference clients from the LLAMA_* and EMBEDDING_*
// environment set by the model profile and retrieval.env. Setting
// INFERENCE_PROVIDER=openai switches chat to an OpenAI-compatible API.
package llamaenv

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/openai"
)

const (
	providerLlama  = "llama"
	providerOpenAI = "openai"
)

// Parallelism is the inference server capacity; the compose default is 1.
func Parallelism() (int, error) {
	value := env(envLlamaParallel)
	if value == "" {
		value = "1"
	}
	parallelism, err := strconv.Atoi(value)
	if err != nil || parallelism < 1 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", envLlamaParallel, value)
	}
	return parallelism, nil
}

// ChatClient is a chat client that reports the metadata of its attempts.
type ChatClient interface {
	inference.Client
	inference.MetadataProvider
}

// SamplingReader is implemented by chat clients whose server exposes its
// sampler defaults (llama-server only).
type SamplingReader interface {
	Sampling(context.Context) (inference.Sampling, error)
}

func NewChatClient() (ChatClient, error) {
	switch provider := env(envInferenceProvider); provider {
	case "", providerLlama:
		return newLlamaChatClient()
	case providerOpenAI:
		return newOpenAIChatClient()
	default:
		return nil, fmt.Errorf("%s must be %q or %q, got %q", envInferenceProvider, providerLlama, providerOpenAI, provider)
	}
}

func newLlamaChatClient() (llama.Adapter, error) {
	model := env(envLlamaModelName)
	if model == "" {
		return llama.Adapter{}, fmt.Errorf("%s is required", envLlamaModelName)
	}
	metadata := inference.Metadata{
		Model:        model,
		Artifact:     artifact(envLlamaModelRepository, envLlamaModelRevision, envLlamaModelFile),
		Quantization: env(envLlamaModelQuant),
		SHA256:       env(envLlamaModelSHA256),
	}
	// A server at LLAMA_BASE_URL is started elsewhere, so the LLAMA_* runtime
	// variables do not describe it.
	if baseURL := env(envLlamaBaseURL); baseURL != "" {
		return newAdapter(baseURL, metadata)
	}
	port := env(envLlamaPort)
	if port == "" {
		port = "8080"
	}
	metadata.RuntimeSettings = runtimeSettings()
	return newAdapter(localURL(port), metadata)
}

func newOpenAIChatClient() (ChatClient, error) {
	baseURL, model := env(envOpenAIBaseURL), env(envOpenAIModel)
	if baseURL == "" || model == "" {
		return nil, fmt.Errorf("%s and %s are required when %s=%s", envOpenAIBaseURL, envOpenAIModel, envInferenceProvider, providerOpenAI)
	}
	client, err := openai.NewClient(openai.Config{
		BaseURL: baseURL,
		Model:   model,
		APIKey:  env(envOpenAIAPIKey),
	})
	if err != nil {
		return nil, fmt.Errorf("create openai client for %s: %w", model, err)
	}
	return client, nil
}

func NewEmbeddingClient() (llama.Adapter, error) {
	model, port, baseURL := env(envEmbeddingModelName), env(envEmbeddingPort), env(envEmbeddingBaseURL)
	if model == "" || port == "" && baseURL == "" {
		return llama.Adapter{}, fmt.Errorf("%s and %s or %s are required", envEmbeddingModelName, envEmbeddingPort, envEmbeddingBaseURL)
	}
	if baseURL == "" {
		baseURL = localURL(port)
	}
	return newAdapter(baseURL, inference.Metadata{
		Model:        model,
		Artifact:     artifact(envEmbeddingRepository, envEmbeddingRevision, envEmbeddingFile),
		Quantization: env(envEmbeddingQuant),
		SHA256:       env(envEmbeddingSHA256),
	})
}

func localURL(port string) string {
	host := env(envLlamaClientHost)
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

func newAdapter(baseURL string, metadata inference.Metadata) (llama.Adapter, error) {
	client, err := llama.NewClient(llama.Config{
		BaseURL:  baseURL,
		Model:    metadata.Model,
		Metadata: metadata,
	})
	if err != nil {
		return llama.Adapter{}, fmt.Errorf("create llama client for %s: %w", metadata.Model, err)
	}
	return llama.NewAdapter(client)
}

// artifact formats repository@revision/file, omitting unset parts.
func artifact(repositoryEnv, revisionEnv, fileEnv string) string {
	result := env(repositoryEnv)
	if revision := env(revisionEnv); revision != "" {
		result += "@" + revision
	}
	if file := env(fileEnv); file != "" {
		result += "/" + file
	}
	return result
}

func runtimeSettings() map[string]string {
	settings := make(map[string]string)
	for _, name := range llamaRuntimeEnvNames {
		if value := env(name); value != "" {
			settings[name] = value
		}
	}
	if len(settings) == 0 {
		return nil
	}
	return settings
}

func env(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}
