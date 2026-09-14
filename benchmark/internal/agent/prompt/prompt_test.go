package prompt

import (
	"context"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type promptTestClient struct {
	requests [][]inference.Message
}

func (c *promptTestClient) Chat(_ context.Context, messages []inference.Message, _ []inference.Tool, _ inference.Options) (inference.Result, error) {
	c.requests = append(c.requests, append([]inference.Message(nil), messages...))
	return inference.Result{
		Message:      inference.Message{Role: "assistant", Content: "done"},
		FinishReason: "stop",
	}, nil
}

type promptTestShell struct{}

func (promptTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{ExitCode: 0}, nil
}

func TestRunUsesVerboseGenericSystemPrompt(t *testing.T) {
	client := &promptTestClient{}
	agent, err := New(client, promptTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the service.")
	require.NoError(t, err)
	assert.Equal(t, "prompt", result.Condition)
	assert.Equal(t, "Restore the service.", result.Task)
	require.Len(t, client.requests, 1)
	require.Len(t, client.requests[0], 2)
	assert.Equal(t, "system", client.requests[0][0].Role)
	assert.Equal(t, VerboseSystemPrompt, client.requests[0][0].Content)
	assert.Equal(t, "user", client.requests[0][1].Role)

	assert.Greater(t, len(VerboseSystemPrompt), len(ShortSystemPrompt)*2)
	assert.Contains(t, VerboseSystemPrompt, "Use Bash deliberately")
	assert.Contains(t, VerboseSystemPrompt, "Use kubectl deliberately")
	assert.Contains(t, VerboseSystemPrompt, "evidence-first troubleshooting loop")
	for _, forbidden := range []string{
		"For workload problems",
		"For service problems",
		"For scheduling problems",
		"For permission problems",
		"ImagePullBackOff",
	} {
		assert.NotContains(t, VerboseSystemPrompt, forbidden)
	}
}

func TestRunRejectsUninitializedAgent(t *testing.T) {
	var nilAgent *Agent
	_, err := nilAgent.Run(context.Background(), "task")
	assert.EqualError(t, err, "prompt agent is not initialized")
}
