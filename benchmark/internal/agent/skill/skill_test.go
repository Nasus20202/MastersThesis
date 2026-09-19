package skill

import (
	"context"
	"testing"

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
	agent, err := New(client, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)

	assert.Equal(t, "skill", result.Condition)
	assert.Equal(t, "Restore the application.", result.Task)
	require.Len(t, client.requests, 1)
	require.Len(t, client.requests[0], 2)
	assert.Equal(t, "system", client.requests[0][0].Role)
	routingPrompt, err := routingSystemPrompt()
	require.NoError(t, err)
	assert.Equal(t, routingPrompt, client.requests[0][0].Content)
	assert.Contains(t, client.requests[0][0].Content, "Use Bash to inspect the sandbox")
	assert.Contains(t, client.requests[0][0].Content, "Use `load_skill` to load only skills useful for the current task")
	assert.Contains(t, client.requests[0][0].Content, "load a listed reference only if more detail is needed")
	assert.Contains(t, client.requests[0][0].Content, "Available skills:")
	assert.NotContains(t, client.requests[0][0].Content, "Identify the requested outcome and limits")
	assert.Equal(t, "user", client.requests[0][1].Role)
	assert.Len(t, client.tools, 3)
	assert.Equal(t, "bash", client.tools[0].Name)
	assert.Equal(t, loadSkillToolName, client.tools[1].Name)
	assert.Equal(t, loadReferenceToolName, client.tools[2].Name)
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
	agent, err := New(client, skillTestShell{}, common.Config{MaxTurns: 3, MaxToolCalls: 2})
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

func TestNewAndRunRejectInvalidSkillAgent(t *testing.T) {
	var nilAgent *Agent
	_, err := nilAgent.Run(context.Background(), "task")
	assert.EqualError(t, err, "skill agent is not initialized")

	_, err = NewWithSystemPrompt(&skillTestClient{}, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, " ")
	assert.EqualError(t, err, "skill system prompt is required")

	agent, err := New(&skillTestClient{}, skillTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)
	_, err = agent.Run(context.Background(), " ")
	assert.EqualError(t, err, "agent task is required")
}
