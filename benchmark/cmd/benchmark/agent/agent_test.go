package agent

import (
	"context"
	"os"
	"testing"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []Name
	}{
		{name: "default", want: []Name{Baseline, Prompt}},
		{name: "all", values: []string{"all"}, want: []Name{Baseline, Prompt}},
		{name: "single", values: []string{"BASELINE"}, want: []Name{Baseline}},
		{name: "trimmed", values: []string{" prompt "}, want: []Name{Prompt}},
		{name: "repeated", values: []string{"baseline", "prompt"}, want: []Name{Baseline, Prompt}},
		{name: "comma-separated", values: []string{"baseline,prompt"}, want: []Name{Baseline, Prompt}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Select(test.values...)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	testsWithErrors := []struct {
		values []string
		want   string
	}{
		{values: []string{"unknown"}, want: "unsupported agent"},
		{values: []string{""}, want: "agent must not be blank"},
		{values: []string{"baseline", "baseline"}, want: "selected more than once"},
		{values: []string{"all", "baseline"}, want: "cannot be combined"},
		{values: []string{"all,prompt"}, want: "cannot be combined"},
	}
	for _, test := range testsWithErrors {
		_, err := Select(test.values...)
		assert.ErrorContains(t, err, test.want)
	}
}

func TestNewBaselineFactoryRequiresModel(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "")

	factory, err := NewBaselineFactory(benchmarkconfig.Config{})
	assert.Nil(t, factory)
	assert.EqualError(t, err, "LLAMA_MODEL_NAME is required")
}

func TestNewFactoryConstructsSupportedAgents(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")
	t.Setenv("LLAMA_CLIENT_HOST", "127.0.0.1")
	t.Setenv("LLAMA_PORT", "8080")

	for _, name := range []Name{Baseline, Prompt} {
		t.Run(string(name), func(t *testing.T) {
			factory, err := NewFactory(name, benchmarkconfig.Config{})
			require.NoError(t, err)
			agent, err := factory(baselineTestExecutor{})
			require.NoError(t, err)
			assert.NotNil(t, agent)
		})
	}
}

func TestNewFactoryRejectsUnsupportedAgent(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")

	factory, err := NewFactory(Name("unknown"), benchmarkconfig.Config{})
	assert.Nil(t, factory)
	assert.ErrorContains(t, err, "unsupported agent")
}

func TestConfiguredLoopConfig(t *testing.T) {
	maxTurns := 12
	maxToolCalls := 20
	toolTimeout := 15.0
	timeout := 90.0
	got, err := configuredLoopConfig(benchmarkconfig.LoopConfig{
		MaxTurns:           &maxTurns,
		MaxToolCalls:       &maxToolCalls,
		ToolTimeoutSeconds: &toolTimeout,
		TimeoutSeconds:     &timeout,
	})
	require.NoError(t, err)
	assert.Equal(t, maxTurns, got.MaxTurns)
	assert.Equal(t, maxToolCalls, got.MaxToolCalls)
	assert.Equal(t, toolTimeout, got.ToolTimeoutSeconds)
	assert.Equal(t, timeout, got.TimeoutSeconds)

	defaultConfig, err := configuredLoopConfig(benchmarkconfig.LoopConfig{})
	require.NoError(t, err)
	assert.Equal(t, 25, defaultConfig.MaxTurns)
	assert.Equal(t, 50, defaultConfig.MaxToolCalls)
	assert.Equal(t, float64(60), defaultConfig.ToolTimeoutSeconds)
	assert.Equal(t, float64(300), defaultConfig.TimeoutSeconds)
}

func TestConfiguredLoopConfigRejectsNonpositiveValues(t *testing.T) {
	zeroInt := 0
	zeroFloat := 0.0
	tests := []struct {
		name   string
		config benchmarkconfig.LoopConfig
	}{
		{name: "max turns", config: benchmarkconfig.LoopConfig{MaxTurns: &zeroInt}},
		{name: "max tool calls", config: benchmarkconfig.LoopConfig{MaxToolCalls: &zeroInt}},
		{name: "tool timeout", config: benchmarkconfig.LoopConfig{ToolTimeoutSeconds: &zeroFloat}},
		{name: "total timeout", config: benchmarkconfig.LoopConfig{TimeoutSeconds: &zeroFloat}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := configuredLoopConfig(test.config)
			assert.Error(t, err)
		})
	}
}

func TestLoadPromptSystemPrompt(t *testing.T) {
	path := t.TempDir() + "/prompt.md"
	require.NoError(t, os.WriteFile(path, []byte("custom instructions\n"), 0o600))
	prompt, err := loadPromptSystemPrompt(path)
	require.NoError(t, err)
	assert.Equal(t, "custom instructions\n", prompt)
}

func TestLoadPromptSystemPromptRejectsMissingAndBlankFiles(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		path := t.TempDir() + "/missing.md"
		_, err := loadPromptSystemPrompt(path)
		assert.ErrorContains(t, err, "read prompt system prompt")
	})

	t.Run("blank", func(t *testing.T) {
		path := t.TempDir() + "/prompt.md"
		require.NoError(t, os.WriteFile(path, []byte(" \n"), 0o600))
		_, err := loadPromptSystemPrompt(path)
		assert.ErrorContains(t, err, "prompt system prompt file must not be blank")
	})
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
