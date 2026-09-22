// Package baseline implements the baseline condition: a bash-only tool loop
// with no system prompt.
package baseline

import (
	"context"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Agent is the baseline condition. It owns the condition's construction while
// delegating the shared model/tool mechanics to common.Agent.
type Agent struct {
	delegate rootagent.Agent
}

const promptInstruction = "You are working in a Kubernetes troubleshooting environment. Use the bash tool to inspect and modify the environment as needed. Complete the task using only the available environment. When finished, provide a short final response."

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

func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	initialized := a != nil && a.delegate != nil
	return common.RunAgent(initialized, "baseline", "baseline", task, func() (common.Result, error) {
		return a.delegate.Run(ctx, baselinePrompt(task))
	})
}

func baselinePrompt(task string) string {
	return "Task:\n" + task + "\n\n" + promptInstruction
}

var _ rootagent.Agent = (*Agent)(nil)
