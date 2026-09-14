package skill

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	promptagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/prompt"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	projectskills "github.com/Nasus20202/MastersThesis/benchmark/internal/skills"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type skillTestClient struct {
	requests [][]inference.Message
	tools    [][]inference.Tool
}

func (c *skillTestClient) Chat(_ context.Context, messages []inference.Message, tools []inference.Tool, _ inference.Options) (inference.Result, error) {
	c.requests = append(c.requests, append([]inference.Message(nil), messages...))
	c.tools = append(c.tools, append([]inference.Tool(nil), tools...))
	return inference.Result{
		Message:      inference.Message{Role: "assistant", Content: "done"},
		FinishReason: "stop",
	}, nil
}

type skillTestShell struct{}

func (skillTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{ExitCode: 0}, nil
}

func newSkillTestRegistry(t *testing.T) *projectskills.Registry {
	t.Helper()
	root := t.TempDir()
	writeSkillTestFile(t, filepath.Join(root, "bash", "SKILL.md"), `---
name: bash
description: Shell guidance.
metadata:
  version: "1"
---

# Bash body
`)
	writeSkillTestFile(t, filepath.Join(root, "kubernetes", "SKILL.md"), `---
name: kubernetes
description: Kubernetes guidance.
metadata:
  version: "1"
---

# Kubernetes body
`)
	writeSkillTestFile(t, filepath.Join(root, "kubernetes", "references", "pods.md"), "# Pods reference\n\nInspect status and events.\n")
	registry, err := projectskills.Load(root)
	require.NoError(t, err)
	return registry
}

func writeSkillTestFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestRunUsesShortManifestDrivenSystemPrompt(t *testing.T) {
	registry := newSkillTestRegistry(t)
	client := &skillTestClient{}
	agent, err := New(client, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, registry)
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Inspect the environment.")
	require.NoError(t, err)
	assert.Equal(t, "skill", result.Condition)
	require.Len(t, client.requests, 1)
	require.Len(t, client.requests[0], 2)
	systemPrompt := client.requests[0][0].Content
	assert.Contains(t, systemPrompt, "benchmark sandbox")
	assert.Contains(t, systemPrompt, "- bash: Shell guidance.")
	assert.Contains(t, systemPrompt, "- kubernetes: Kubernetes guidance.")
	assert.Contains(t, systemPrompt, "references: pods.md")
	assert.Contains(t, systemPrompt, "Use load_skill")
	assert.NotContains(t, systemPrompt, "# Kubernetes body")
	assert.NotContains(t, systemPrompt, "evidence-first troubleshooting loop")
	assert.Less(t, len(systemPrompt), len(promptagent.VerboseSystemPrompt))

	require.Len(t, client.tools, 1)
	assert.Equal(t, []string{"bash", "load_skill"}, []string{client.tools[0][0].Name, client.tools[0][1].Name})
}

func TestLoadSkillToolReturnsMainAndReferenceContent(t *testing.T) {
	registry := newSkillTestRegistry(t)
	tool, err := NewLoadSkillTool(registry)
	require.NoError(t, err)
	assert.True(t, json.Valid(tool.Definition().Parameters))

	main := tool.Execute(context.Background(), inference.ToolCall{
		Type:      "function",
		Name:      "load_skill",
		Arguments: `{"skill":"kubernetes"}`,
	})
	require.NoError(t, main.Error)
	assert.Contains(t, main.Content, "Skill: kubernetes")
	assert.Contains(t, main.Content, "# Kubernetes body")
	mainEvidence, ok := main.Details.(SkillLoadEvidence)
	require.True(t, ok)
	assert.Equal(t, "SKILL.md", mainEvidence.Reference)

	reference := tool.Execute(context.Background(), inference.ToolCall{
		Type:      "function",
		Name:      "load_skill",
		Arguments: `{"skill":"kubernetes","reference":"pods.md"}`,
	})
	require.NoError(t, reference.Error)
	assert.Contains(t, reference.Content, "# Pods reference")
	referenceEvidence, ok := reference.Details.(SkillLoadEvidence)
	require.True(t, ok)
	assert.Equal(t, "pods.md", referenceEvidence.Reference)

	invalid := tool.Execute(context.Background(), inference.ToolCall{
		Type:      "function",
		Name:      "load_skill",
		Arguments: `{"skill":"kubernetes","reference":"../SKILL.md"}`,
	})
	assert.ErrorContains(t, invalid.Error, "invalid")
}

func TestNewRejectsMissingRegistry(t *testing.T) {
	_, err := New(&skillTestClient{}, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, nil)
	assert.EqualError(t, err, "skills registry is required")
}
