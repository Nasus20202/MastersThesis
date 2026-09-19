package skill

import (
	"context"
	"embed"
	"errors"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

//go:embed prompt.md
var systemPrompt string

//go:embed skills
var skillFiles embed.FS

type Agent struct {
	loop         *common.Loop
	systemPrompt string
}

func New(client inference.Client, shell common.Shell, config common.Config) (*Agent, error) {
	prompt, err := routingSystemPrompt()
	if err != nil {
		return nil, err
	}
	return NewWithSystemPrompt(client, shell, config, prompt)
}

func NewWithSystemPrompt(client inference.Client, shell common.Shell, config common.Config, systemPrompt string) (*Agent, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		return nil, errors.New("skill system prompt is required")
	}
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	loader := newSkillTool()
	referenceLoader := newReferenceTool()
	loop, err := common.NewLoop(client, []common.Tool{bash, loader, referenceLoader}, config)
	if err != nil {
		return nil, err
	}
	return &Agent{loop: loop, systemPrompt: rootagent.WithExecutionContract(systemPrompt)}, nil
}

func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	if a == nil || a.loop == nil {
		return common.Result{}, errors.New("skill agent is not initialized")
	}
	if strings.TrimSpace(task) == "" {
		return common.Result{}, errors.New("agent task is required")
	}
	result, err := a.loop.RunWithSystemPrompt(ctx, task, a.systemPrompt)
	result.Condition = "skill"
	result.Task = task
	return result, err
}

var _ rootagent.Agent = (*Agent)(nil)
