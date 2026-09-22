// Package skill implements the skill condition: a tool loop that can
// discover and load Kubernetes troubleshooting skills.
package skill

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
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
	return newAgent(client, shell, config, systemPrompt, skillFiles)
}

// NewWithConfig uses configured files when supplied and falls back to the
// embedded skill bundle for omitted paths.
func NewWithConfig(client inference.Client, shell common.Shell, config common.Config, systemPromptFile, skillsDir string) (*Agent, error) {
	prompt := systemPrompt
	if strings.TrimSpace(systemPromptFile) != "" {
		data, err := os.ReadFile(systemPromptFile)
		if err != nil {
			return nil, fmt.Errorf("read skill system prompt: %w", err)
		}
		prompt = string(data)
	}
	files := fs.FS(skillFiles)
	if strings.TrimSpace(skillsDir) != "" {
		files = os.DirFS(skillsDir)
	}
	return newAgent(client, shell, config, prompt, files)
}

func newAgent(client inference.Client, shell common.Shell, config common.Config, basePrompt string, files fs.FS) (*Agent, error) {
	prompt, err := routingSystemPromptFor(basePrompt, files)
	if err != nil {
		return nil, err
	}
	return NewWithSystemPromptAndFiles(client, shell, config, prompt, files)
}

func NewWithSystemPrompt(client inference.Client, shell common.Shell, config common.Config, systemPrompt string) (*Agent, error) {
	return NewWithSystemPromptAndFiles(client, shell, config, systemPrompt, skillFiles)
}

func NewWithSystemPromptAndFiles(client inference.Client, shell common.Shell, config common.Config, systemPrompt string, files fs.FS) (*Agent, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		return nil, errors.New("skill system prompt is required")
	}
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	loader := newSkillTool(files)
	referenceLoader := newReferenceTool(files)
	loop, err := common.NewLoop(client, []common.Tool{bash, loader, referenceLoader}, config)
	if err != nil {
		return nil, err
	}
	return &Agent{loop: loop, systemPrompt: systemPrompt}, nil
}

func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	initialized := a != nil && a.loop != nil
	return common.RunAgent(initialized, "skill", "skill", task, func() (common.Result, error) {
		return a.loop.RunWithSystemPrompt(ctx, task, a.systemPrompt)
	})
}

var _ rootagent.Agent = (*Agent)(nil)
