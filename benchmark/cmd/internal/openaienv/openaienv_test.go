package openaienv

import (
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChatClient(t *testing.T) {
	t.Setenv(envBaseURL, "")
	t.Setenv(envModel, "")
	_, err := NewChatClient()
	assert.ErrorContains(t, err, envBaseURL+" and "+envModel+" are required")

	t.Setenv(envBaseURL, "https://api.example.test/v1/")
	t.Setenv(envModel, "gpt-test")
	client, err := NewChatClient()
	require.NoError(t, err)
	assert.Equal(t, inference.Metadata{Provider: "openai", Model: "gpt-test"}, client.Metadata())

	t.Setenv(envEffort, "low")
	client, err = NewChatClient()
	require.NoError(t, err)
	assert.Equal(t, "low", client.Metadata().RuntimeSettings["reasoning_effort"])
}
