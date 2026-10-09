package inferenceenv

import (
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChatClientProvider(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma")
	t.Setenv("OPENAI_BASE_URL", "https://api.example.test/v1")
	t.Setenv("OPENAI_MODEL", "gpt-test")

	for provider, wantSampling := range map[provider]bool{"": true, providerLlama: true, providerOpenAI: false} {
		t.Run("provider "+string(provider), func(t *testing.T) {
			t.Setenv(envProvider, string(provider))
			client, err := NewChatClient()
			require.NoError(t, err)
			_, hasSampling := client.(inference.SamplingReader)
			assert.Equal(t, wantSampling, hasSampling)
		})
	}

	t.Setenv(envProvider, "other")
	_, err := NewChatClient()
	assert.ErrorContains(t, err, envProvider+" must be")
}
