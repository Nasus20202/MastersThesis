package agent

import (
	"context"
	"os"
	"testing"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	ragagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/rag"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []Name
	}{
		{name: "default", want: []Name{Baseline, Prompt, Skill, RAG}},
		{name: "all", values: []string{"all"}, want: []Name{Baseline, Prompt, Skill, RAG}},
		{name: "rag", values: []string{"RAG"}, want: []Name{RAG}},
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

func TestNewFactoryRequiresModel(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "")

	factory, err := NewFactory(Baseline, benchmarkconfig.Config{}, nil)
	assert.Nil(t, factory)
	assert.EqualError(t, err, "LLAMA_MODEL_NAME is required")
}

func TestNewFactoryConstructsSupportedAgents(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")
	t.Setenv("LLAMA_CLIENT_HOST", "127.0.0.1")
	t.Setenv("LLAMA_PORT", "8080")

	for _, name := range []Name{Baseline, Prompt, Skill} {
		t.Run(string(name), func(t *testing.T) {
			factory, err := NewFactory(name, benchmarkconfig.Config{}, nil)
			require.NoError(t, err)
			agent, err := factory(baselineTestExecutor{})
			require.NoError(t, err)
			assert.NotNil(t, agent)
		})
	}
}

func TestNewFactoryRAGRequiresSearch(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")

	factory, err := NewFactory(RAG, benchmarkconfig.Config{}, nil)
	assert.Nil(t, factory)
	assert.EqualError(t, err, "rag agent requires an opened retrieval index")

	ctx := context.Background()
	path := retrieval.IndexPath(t.TempDir(), retrieval.Windows)
	_, err = retrieval.BuildIndex(ctx, path, []retrieval.Document{{Path: "docs/a.md", Title: "A", Text: "Body"}},
		retrieval.Windows, unitEmbedder{}, retrieval.IndexMetadata{})
	require.NoError(t, err)
	index, err := retrieval.OpenIndex(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = index.Close() })

	factory, err = NewFactory(RAG, benchmarkconfig.Config{}, &ragagent.Search{Index: index, Embedder: unitEmbedder{}, Mode: retrieval.Lexical, TopK: 5, MaxBytes: 8192})
	require.NoError(t, err)
	agent, err := factory(baselineTestExecutor{})
	require.NoError(t, err)
	assert.NotNil(t, agent)
}

func TestOpenSearchValidatesSettings(t *testing.T) {
	_, _, err := OpenSearch(context.Background(), benchmarkconfig.RetrievalConfig{})
	assert.ErrorContains(t, err, "retrieval.top_k, retrieval.max_bytes and retrieval.index_dir are required")
}

type unitEmbedder struct{}

func (unitEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for index := range embeddings {
		embeddings[index] = []float32{1, 0}
	}
	return embeddings, nil
}

func TestNewFactoryRejectsUnsupportedAgent(t *testing.T) {
	t.Setenv("LLAMA_MODEL_NAME", "gemma-test")

	factory, err := NewFactory(Name("unknown"), benchmarkconfig.Config{}, nil)
	assert.Nil(t, factory)
	assert.ErrorContains(t, err, "unsupported agent")
}

func TestConfiguredLoopConfig(t *testing.T) {
	maxTurns := 12
	maxToolCalls := 20
	toolTimeout := 15.0
	timeout := 90.0
	fileTools := false
	got, err := configuredLoopConfig(benchmarkconfig.LoopConfig{
		MaxTurns:           &maxTurns,
		MaxToolCalls:       &maxToolCalls,
		ToolTimeoutSeconds: &toolTimeout,
		TimeoutSeconds:     &timeout,
		FileTools:          &fileTools,
	})
	require.NoError(t, err)
	assert.Equal(t, maxTurns, got.MaxTurns)
	assert.Equal(t, maxToolCalls, got.MaxToolCalls)
	assert.Equal(t, toolTimeout, got.ToolTimeoutSeconds)
	assert.Equal(t, timeout, got.TimeoutSeconds)
	assert.False(t, got.FileTools)

	defaultConfig, err := configuredLoopConfig(benchmarkconfig.LoopConfig{})
	require.NoError(t, err)
	assert.Equal(t, 25, defaultConfig.MaxTurns)
	assert.Equal(t, 50, defaultConfig.MaxToolCalls)
	assert.Equal(t, float64(60), defaultConfig.ToolTimeoutSeconds)
	assert.Equal(t, float64(300), defaultConfig.TimeoutSeconds)
	assert.True(t, defaultConfig.FileTools)
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

func TestReadSystemPrompt(t *testing.T) {
	prompt, err := readSystemPrompt(" ")
	require.NoError(t, err)
	assert.Empty(t, prompt)

	path := t.TempDir() + "/prompt.md"
	require.NoError(t, os.WriteFile(path, []byte("custom instructions\n"), 0o600))
	prompt, err = readSystemPrompt(path)
	require.NoError(t, err)
	assert.Equal(t, "custom instructions\n", prompt)

	_, err = readSystemPrompt(t.TempDir() + "/missing.md")
	assert.ErrorContains(t, err, "read system prompt")

	require.NoError(t, os.WriteFile(path, []byte(" \n"), 0o600))
	_, err = readSystemPrompt(path)
	assert.ErrorContains(t, err, "must not be blank")
}

type baselineTestExecutor struct{}

func (baselineTestExecutor) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}
