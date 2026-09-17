package agent

import (
	"context"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	tests := []struct {
		value string
		want  []Name
	}{
		{value: "all", want: []Name{Baseline, Prompt}},
		{value: "BASELINE", want: []Name{Baseline}},
		{value: " prompt ", want: []Name{Prompt}},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, err := Select(test.value)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	_, err := Select("unknown")
	assert.ErrorContains(t, err, "unsupported agent")
}

func TestNewBaselineFactoryRequiresModel(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "")

	factory, err := NewBaselineFactory()
	assert.Nil(t, factory)
	assert.EqualError(t, err, "LLAMA_MODEL_NAME is required")
}

func TestNewFactoryConstructsSupportedAgents(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")
	t.Setenv("LLAMA_CLIENT_HOST", "127.0.0.1")
	t.Setenv("LLAMA_PORT", "8080")

	for _, name := range []Name{Baseline, Prompt} {
		t.Run(string(name), func(t *testing.T) {
			factory, err := NewFactory(name)
			require.NoError(t, err)
			agent, err := factory(baselineTestExecutor{})
			require.NoError(t, err)
			assert.NotNil(t, agent)
		})
	}
}

func TestNewFactoryRejectsUnsupportedAgent(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")

	factory, err := NewFactory(Name("unknown"))
	assert.Nil(t, factory)
	assert.ErrorContains(t, err, "unsupported agent")
}

func TestModelArtifactAndRuntimeSettings(t *testing.T) {
	t.Setenv("LLAMA_MODEL_REPOSITORY", "google/gemma")
	t.Setenv("LLAMA_MODEL_REVISION", "revision")
	t.Setenv("LLAMA_MODEL_FILE", "gemma.gguf")
	t.Setenv("LLAMA_MODEL_QUANTIZATION", "Q4_0")
	t.Setenv("LLAMA_MODEL_SHA256", "hash")
	t.Setenv("LLAMA_CONTEXT_SIZE", "32768")
	t.Setenv("LLAMA_FLASH_ATTN", "auto")
	t.Setenv("LLAMA_REASONING", "on")

	assert.Equal(t, "google/gemma@revision/gemma.gguf", modelArtifact())
	settings := runtimeSettings()
	assert.Equal(t, "32768", settings["LLAMA_CONTEXT_SIZE"])
	assert.Equal(t, "on", settings["LLAMA_REASONING"])
	assert.Equal(t, "auto", settings["LLAMA_FLASH_ATTN"])
}

type baselineTestExecutor struct{}

func (baselineTestExecutor) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}
