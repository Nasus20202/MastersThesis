package agent

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaselineFactoryRequiresModel(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "")

	factory, err := NewBaselineFactory()
	assert.Nil(t, factory)
	assert.EqualError(t, err, "LLAMA_MODEL_NAME is required")
}

func TestNewBaselineFactoryConstructsAgent(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")
	t.Setenv("LLAMA_CLIENT_HOST", "127.0.0.1")
	t.Setenv("LLAMA_PORT", "8080")

	factory, err := NewBaselineFactory()
	require.NoError(t, err)
	agent, err := factory(baselineTestExecutor{})
	require.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestResolveNamesUsesAllAgentsByDefaultAndPreservesSelectionOrder(t *testing.T) {
	names, err := ResolveNames(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{string(Baseline), string(Prompt), string(Skill)}, names)

	names, err = ResolveNames([]string{string(Skill), string(Baseline)})
	require.NoError(t, err)
	assert.Equal(t, []string{string(Skill), string(Baseline)}, names)
}

func TestResolveNamesRejectsUnknownAndDuplicateAgents(t *testing.T) {
	_, err := ResolveNames([]string{"unknown"})
	assert.ErrorContains(t, err, "unsupported agent")
	_, err = ResolveNames([]string{string(Baseline), string(Baseline)})
	assert.ErrorContains(t, err, "selected more than once")
}

func TestNewFactoriesConstructsAllSelectedAgents(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")
	t.Setenv("BENCHMARK_SKILLS_DIR", "")
	skillsRoot, err := skillsDirectory()
	require.NoError(t, err)
	t.Setenv("BENCHMARK_SKILLS_DIR", filepath.Clean(skillsRoot))

	factories, err := NewFactories(nil)
	require.NoError(t, err)
	assert.Len(t, factories, 3)
	for _, name := range AllNames() {
		factory := factories[name]
		require.NotNil(t, factory)
		modelAgent, err := factory(baselineTestExecutor{})
		require.NoError(t, err)
		assert.NotNil(t, modelAgent)
	}
}

func TestNewFactoriesLoadsSkillsOnlyWhenSelected(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")
	t.Setenv("BENCHMARK_SKILLS_DIR", filepath.Join(t.TempDir(), "missing"))

	factories, err := NewFactories([]string{string(Baseline)})
	require.NoError(t, err)
	assert.Len(t, factories, 1)

	_, err = NewFactories([]string{string(Skill)})
	assert.ErrorContains(t, err, "load benchmark skills")
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
