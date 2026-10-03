package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type skillTestClient struct {
	results  []inference.Result
	requests [][]inference.Message
	tools    []inference.Tool
}

func (c *skillTestClient) Chat(_ context.Context, messages []inference.Message, tools []inference.Tool, _ inference.Options) (inference.Result, error) {
	request := append([]inference.Message(nil), messages...)
	c.requests = append(c.requests, request)
	c.tools = append([]inference.Tool(nil), tools...)
	index := len(c.requests) - 1
	if index < len(c.results) {
		return c.results[index], nil
	}
	return inference.Result{Message: inference.Message{Role: "assistant", Content: "Done."}, FinishReason: "stop"}, nil
}

type skillTestShell struct{}

func (skillTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}

func TestRunUsesRoutingPromptAndSkillLoader(t *testing.T) {
	client := &skillTestClient{}
	agent, err := New(client, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, "", nil)
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)

	assert.Equal(t, "skill", result.Condition)
	assert.Equal(t, "Restore the application.", result.Task)
	require.Len(t, client.requests, 1)
	require.Len(t, client.requests[0], 2)
	assert.Equal(t, "system", client.requests[0][0].Role)
	routingPrompt, err := routingSystemPrompt(defaultSystemPrompt, defaultSkillFiles)
	require.NoError(t, err)
	assert.Equal(t, routingPrompt, client.requests[0][0].Content)
	assert.Contains(t, client.requests[0][0].Content, "Use Bash to inspect the sandbox")
	assert.Contains(t, client.requests[0][0].Content, "Before the first Bash call, load the `troubleshooting` skill")
	assert.Contains(t, client.requests[0][0].Content, "Match the observed symptom to the subsystem")
	assert.Contains(t, client.requests[0][0].Content, "load the reference that covers the affected subsystem with `load_reference`")
	assert.Contains(t, client.requests[0][0].Content, "using `skill: \"kubernetes\"` and the exact filename")
	assert.Contains(t, client.requests[0][0].Content, "Load `kubectl` as well when command or patch semantics are uncertain")
	assert.Contains(t, client.requests[0][0].Content, "Prove the requested outcome before reporting success")
	assert.Contains(t, client.requests[0][0].Content, "Report `Outcome: complete` only when every required check passes")
	assert.NotContains(t, client.requests[0][0].Content, "Load other skills and listed references only when their guidance is useful")
	assert.Contains(t, client.requests[0][0].Content, "Available skills:")
	assert.NotContains(t, client.requests[0][0].Content, "Identify the requested outcome and limits")
	assert.Equal(t, "user", client.requests[0][1].Role)
	assert.Len(t, client.tools, 3)
	assert.Equal(t, "bash", client.tools[0].Name)
	assert.Equal(t, loadSkillToolName, client.tools[1].Name)
	assert.Equal(t, loadReferenceToolName, client.tools[2].Name)
}

func TestNewUsesConfiguredPromptAndSkillDirectory(t *testing.T) {
	directory := t.TempDir()
	skillsDirectory := filepath.Join(directory, "skills")
	customSkillDirectory := filepath.Join(skillsDirectory, "custom")
	require.NoError(t, os.MkdirAll(customSkillDirectory, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(customSkillDirectory, "SKILL.md"), []byte("---\nname: custom\ndescription: Custom guidance\n---\n# Custom\n"), 0o600))

	client := &skillTestClient{}
	agent, err := New(client, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, "custom routing instructions", os.DirFS(skillsDirectory))
	require.NoError(t, err)
	_, err = agent.Run(context.Background(), "task")
	require.NoError(t, err)
	require.Len(t, client.requests, 1)
	assert.Contains(t, client.requests[0][0].Content, "custom routing instructions")
	assert.Contains(t, client.requests[0][0].Content, "- custom: Custom guidance")
	assert.NotContains(t, client.requests[0][0].Content, "- kubernetes:")
}

func TestRunLoadsSkillThenReferenceAcrossTurns(t *testing.T) {
	client := &skillTestClient{results: []inference.Result{
		{
			Message: inference.Message{Role: "assistant", ToolCalls: []inference.ToolCall{{
				ID: "skill-1", Type: "function", Name: loadSkillToolName, Arguments: `{"name":"kubernetes"}`,
			}}},
			FinishReason: "tool_calls",
		},
		{
			Message: inference.Message{Role: "assistant", ToolCalls: []inference.ToolCall{{
				ID: "reference-1", Type: "function", Name: loadReferenceToolName, Arguments: `{"skill":"kubernetes","reference":"networking.md"}`,
			}}},
			FinishReason: "tool_calls",
		},
		{Message: inference.Message{Role: "assistant", Content: "Done."}, FinishReason: "stop"},
	}}
	agent, err := New(client, skillTestShell{}, common.Config{MaxTurns: 3, MaxToolCalls: 2}, "", nil)
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)

	require.Len(t, client.requests, 3)
	assert.Contains(t, client.requests[1][len(client.requests[1])-1].Content, "# Kubernetes")
	assert.Contains(t, client.requests[2][len(client.requests[2])-1].Content, "# Kubernetes Networking")
	require.Len(t, result.ToolCalls, 2)
	assert.Equal(t, skillEvidence{Skill: "kubernetes"}, result.ToolCalls[0].Details)
	assert.Equal(t, skillEvidence{Skill: "kubernetes", Reference: "networking.md"}, result.ToolCalls[1].Details)
	assert.Equal(t, common.TerminationCompleted, result.Termination)
}

func TestRunRejectsBlankTask(t *testing.T) {
	agent, err := New(&skillTestClient{}, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, "", nil)
	require.NoError(t, err)
	_, err = agent.Run(context.Background(), " ")
	assert.EqualError(t, err, "agent task is required")
}

func TestChangeHintAppendsKubectlSectionOncePerAttempt(t *testing.T) {
	section := changeSection(defaultSkillFiles)
	require.NotEmpty(t, section)
	assert.Contains(t, section, "strategic merge patch")
	assert.NotContains(t, section, "\n## ")

	hint := &changeHint{section: section}
	text, details := hint.hint(context.Background(), common.FailedChange{Verb: "patch"})
	assert.Equal(t, "Guidance from the kubectl skill on changing state:\n"+section, text)
	assert.Equal(t, ChangeHintEvidence{Skill: "kubectl", Section: "Changing state"}, details)

	text, details = hint.hint(context.Background(), common.FailedChange{Verb: "patch"})
	assert.Empty(t, text)
	assert.Nil(t, details)
}

func TestChangeSectionIsEmptyWithoutKubectlSkill(t *testing.T) {
	assert.Empty(t, changeSection(fstest.MapFS{"custom/SKILL.md": {Data: []byte("# Custom\n")}}))
	assert.Empty(t, changeSection(fstest.MapFS{"kubectl/SKILL.md": {Data: []byte("---\nname: kubectl\ndescription: d\n---\n# kubectl\n\n## Reading\n")}}))
}
