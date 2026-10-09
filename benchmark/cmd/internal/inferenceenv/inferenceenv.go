// Package inferenceenv selects the chat provider named by INFERENCE_PROVIDER.
package inferenceenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/openaienv"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const envProvider = "INFERENCE_PROVIDER"

type provider string

const (
	providerLlama  provider = "llama"
	providerOpenAI provider = "openai"
)

// NewChatClient returns the llama-server client unless INFERENCE_PROVIDER is
// "openai".
func NewChatClient() (inference.ChatClient, error) {
	switch selected := provider(strings.TrimSpace(os.Getenv(envProvider))); selected {
	case "", providerLlama:
		return llamaenv.NewChatClient()
	case providerOpenAI:
		return openaienv.NewChatClient()
	default:
		return nil, fmt.Errorf("%s must be %q or %q, got %q", envProvider, providerLlama, providerOpenAI, selected)
	}
}
