// Package openaienv creates the OpenAI-compatible chat client from the OPENAI_*
// environment.
package openaienv

import (
	"fmt"
	"os"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/openai"
)

const (
	envBaseURL = "OPENAI_BASE_URL"
	envAPIKey  = "OPENAI_API_KEY" //nolint:gosec // variable name, not a credential
	envModel   = "OPENAI_MODEL"
	envEffort  = "OPENAI_REASONING_EFFORT"
)

func NewChatClient() (*openai.Client, error) {
	baseURL, model := env(envBaseURL), env(envModel)
	if baseURL == "" || model == "" {
		return nil, fmt.Errorf("%s and %s are required", envBaseURL, envModel)
	}
	client, err := openai.NewClient(openai.Config{
		BaseURL:         baseURL,
		Model:           model,
		APIKey:          env(envAPIKey),
		ReasoningEffort: env(envEffort),
	})
	if err != nil {
		return nil, fmt.Errorf("create openai client for %s: %w", model, err)
	}
	return client, nil
}

func env(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}
