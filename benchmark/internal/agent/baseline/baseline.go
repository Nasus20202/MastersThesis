package baseline

import (
	"context"
	"errors"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Agent is the baseline condition. It owns the condition's construction while
// delegating the shared model/tool mechanics to common.Agent.
type Agent struct {
	delegate rootagent.Agent
}

// New constructs a baseline agent with the shared loop and sandbox-backed Bash
// tool.
func New(client inference.Client, shell common.Shell, config common.Config) (*Agent, error) {
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	loop, err := common.NewLoop(client, []common.Tool{bash}, config)
	if err != nil {
		return nil, err
	}
	return &Agent{delegate: loop}, nil
}

// Run executes one baseline model/tool interaction.
func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	if a == nil || a.delegate == nil {
		return common.Result{}, errors.New("baseline agent is not initialized")
	}
	return a.delegate.Run(ctx, task)
}

var _ rootagent.Agent = (*Agent)(nil)
