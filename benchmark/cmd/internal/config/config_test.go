package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLayersFilesAndResolvesAgentPaths(t *testing.T) {
	directory := t.TempDir()
	basePath := filepath.Join(directory, "base.yaml")
	overlayPath := filepath.Join(directory, "overlay.yaml")
	require.NoError(t, os.WriteFile(basePath, []byte("logging:\n  level: info\nagents:\n  loop:\n    max_turns: 25\n    max_tool_calls: 50\n    tool_timeout_seconds: 60\n    timeout_seconds: 300\n  prompt:\n    system_prompt_file: default.md\n  skill:\n    system_prompt_file: default-skill.md\n    skills_dir: skills\n  rag:\n    system_prompt_file: rag.md\n"), 0o600))
	require.NoError(t, os.WriteFile(overlayPath, []byte("logging:\n  format: json\nagents:\n  loop:\n    timeout_seconds: 120\n  prompt:\n    system_prompt_file: candidate.md\n  skill:\n    system_prompt_file: candidate-skill.md\n    skills_dir: candidate-skills\n"), 0o600))

	config, err := Load(basePath, overlayPath)
	require.NoError(t, err)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, 25, *config.Agents.Loop.MaxTurns)
	assert.Equal(t, 50, *config.Agents.Loop.MaxToolCalls)
	assert.Equal(t, float64(60), *config.Agents.Loop.ToolTimeoutSeconds)
	assert.Equal(t, float64(120), *config.Agents.Loop.TimeoutSeconds)
	assert.Equal(t, filepath.Join(directory, "candidate.md"), config.Agents.Prompt.SystemPromptFile)
	assert.Equal(t, filepath.Join(directory, "candidate-skill.md"), config.Agents.Skill.SystemPromptFile)
	assert.Equal(t, filepath.Join(directory, "candidate-skills"), config.Agents.Skill.SkillsDir)
	assert.Equal(t, filepath.Join(directory, "rag.md"), config.Agents.RAG.SystemPromptFile)
}

func TestLoadResolvesRetrievalIndexDir(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("retrieval:\n  mode: hybrid\n  top_k: 5\n  index_dir: corpus\n"), 0o600))

	config, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "hybrid", config.Retrieval.Mode)
	assert.Equal(t, 5, *config.Retrieval.TopK)
	assert.Equal(t, filepath.Join(directory, "corpus"), config.Retrieval.IndexDir)
}

func TestLoadUsesDefaultsWithoutFiles(t *testing.T) {
	config, err := Load()
	require.NoError(t, err)
	assert.Equal(t, Config{}, config)
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("unknown: value\n"), 0o600))

	_, err := Load(path)
	assert.ErrorContains(t, err, "invalid benchmark config")
}

func TestLoadRejectsMultipleDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("logging:\n  level: info\n---\nlogging:\n  level: debug\n"), 0o600))

	_, err := Load(path)
	assert.ErrorContains(t, err, "exactly one document")
}

func TestLoadRejectsMalformedAndMissingFiles(t *testing.T) {
	t.Run("malformed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		require.NoError(t, os.WriteFile(path, []byte("logging: [\n"), 0o600))

		_, err := Load(path)
		assert.ErrorContains(t, err, "invalid benchmark config")
	})

	t.Run("missing", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
		assert.ErrorContains(t, err, "read benchmark config")
	})
}
