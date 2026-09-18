package prompt

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

//go:embed prompt.md
var systemPrompt string

type Agent struct {
	loop         *common.Loop
	systemPrompt string
}

func New(client inference.Client, shell common.Shell, config common.Config) (*Agent, error) {
	return NewWithSystemPrompt(client, shell, config, systemPrompt)
}

func NewWithSystemPrompt(client inference.Client, shell common.Shell, config common.Config, systemPrompt string) (*Agent, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		return nil, errors.New("prompt system prompt is required")
	}
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	loop, err := common.NewLoop(client, []common.Tool{bash}, config)
	if err != nil {
		return nil, err
	}
	return &Agent{loop: loop, systemPrompt: systemPrompt}, nil
}

func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	if a == nil || a.loop == nil {
		return common.Result{}, errors.New("prompt agent is not initialized")
	}
	if strings.TrimSpace(task) == "" {
		return common.Result{}, errors.New("agent task is required")
	}
	result, err := a.loop.RunWithSystemPrompt(ctx, task, a.systemPrompt)
	result.Condition = "prompt"
	result.Task = task
	return result, err
}

var _ rootagent.Agent = (*Agent)(nil)
