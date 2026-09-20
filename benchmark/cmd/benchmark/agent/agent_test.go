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
		{name: "default", want: []Name{Baseline, Prompt, Skill}},
		{name: "all", values: []string{"all"}, want: []Name{Baseline, Prompt, Skill}},
		{name: "single", values: []string{"BASELINE"}, want: []Name{Baseline}},
		{name: "skill", values: []string{"skill"}, want: []Name{Skill}},
		{name: "trimmed", values: []string{" prompt "}, want: []Name{Prompt}},
		{name: "repeated", values: []string{"baseline", "prompt", "skill"}, want: []Name{Baseline, Prompt, Skill}},
		{name: "comma-separated", values: []string{"baseline,prompt,skill"}, want: []Name{Baseline, Prompt, Skill}},
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
	t.Setenv(envLlamaModelName, "")

	factory, err := NewBaselineFactory(benchmarkconfig.Config{})
	assert.Nil(t, factory)
	assert.EqualError(t, err, envLlamaModelName+" is required")
}

func TestNewFactoryConstructsSupportedAgents(t *testing.T) {
	t.Setenv(envLlamaModelName, "gemma-test")
	t.Setenv(envLlamaClientHost, "127.0.0.1")
	t.Setenv(envLlamaPort, "8080")

	for _, name := range []Name{Baseline, Prompt, Skill} {
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
	t.Setenv(envLlamaModelName, "gemma-test")

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

func TestConfiguredLlamaParallelism(t *testing.T) {
	t.Run("configured", func(t *testing.T) {
		t.Setenv(envLlamaParallel, "2")
		parallelism, err := ConfiguredLlamaParallelism()
		require.NoError(t, err)
		assert.Equal(t, 2, parallelism)
	})

	t.Run("compose default", func(t *testing.T) {
		t.Setenv(envLlamaParallel, "")
		parallelism, err := ConfiguredLlamaParallelism()
		require.NoError(t, err)
		assert.Equal(t, 1, parallelism)
	})

	for _, value := range []string{"0", "-1", "many"} {
		t.Run("invalid "+value, func(t *testing.T) {
			t.Setenv(envLlamaParallel, value)
			_, err := ConfiguredLlamaParallelism()
			assert.ErrorContains(t, err, envLlamaParallel+" must be a positive integer")
		})
	}
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
	t.Setenv(envLlamaModelRepository, "google/gemma")
	t.Setenv(envLlamaModelRevision, "revision")
	t.Setenv(envLlamaModelFile, "gemma.gguf")
	t.Setenv(envLlamaModelQuant, "Q4_0")
	t.Setenv(envLlamaModelSHA256, "hash")
	t.Setenv(envLlamaKVUnified, "32768")
	t.Setenv(envLlamaFlashAttention, "auto")
	t.Setenv(envLlamaReasoning, "on")

	assert.Equal(t, "google/gemma@revision/gemma.gguf", modelArtifact())
	settings := runtimeSettings()
	assert.Equal(t, "32768", settings[envLlamaKVUnified])
	assert.Equal(t, "on", settings[envLlamaReasoning])
	assert.Equal(t, "auto", settings[envLlamaFlashAttention])
}

type baselineTestExecutor struct{}

func (baselineTestExecutor) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}
