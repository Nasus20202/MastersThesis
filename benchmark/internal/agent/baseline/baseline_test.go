package baseline

import (
	"context"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
)

func TestNewRejectsMissingDependencies(t *testing.T) {
	_, err := New(nil, nil, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	assert.EqualError(t, err, "bash shell is required")

	_, err = New(nil, &baselineTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	assert.EqualError(t, err, "agent llama client is required")
}

func TestRunRejectsUninitializedAgent(t *testing.T) {
	var nilAgent *Agent
	_, err := nilAgent.Run(context.Background(), "task")
	assert.EqualError(t, err, "baseline agent is not initialized")

	_, err = (&Agent{}).Run(context.Background(), "task")
	assert.EqualError(t, err, "baseline agent is not initialized")
}

type baselineTestShell struct{}

func (baselineTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}
