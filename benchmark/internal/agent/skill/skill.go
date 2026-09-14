// Package skill implements the manifest-driven skill condition.
package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/prompt"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/skills"
)

// Agent is the skill condition. The model receives a short fixed context and
// a manifest-derived catalog, then loads approved skill content on demand.
type Agent struct {
	delegate rootagent.Agent
}

func New(client inference.Client, shell common.Shell, config common.Config, registry *skills.Registry) (*Agent, error) {
	if registry == nil {
		return nil, errors.New("skills registry is required")
	}
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	loadSkill, err := NewLoadSkillTool(registry)
	if err != nil {
		return nil, err
	}
	systemPrompt := strings.Join([]string{
		prompt.ShortSystemPrompt,
		"Available skills from their manifests:\n" + registry.Catalog(),
		"Use load_skill with a skill name and, when needed, one listed reference filename. Load only the knowledge relevant to the task.",
	}, "\n\n")
	loop, err := common.NewLoopWithSystemPrompt(client, []common.Tool{bash, loadSkill}, config, systemPrompt)
	if err != nil {
		return nil, err
	}
	return &Agent{delegate: loop}, nil
}

func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	if a == nil || a.delegate == nil {
		return common.Result{}, errors.New("skill agent is not initialized")
	}
	if strings.TrimSpace(task) == "" {
		return common.Result{}, errors.New("agent task is required")
	}
	result, err := a.delegate.Run(ctx, task)
	result.Condition = "skill"
	result.Task = task
	return result, err
}

var _ rootagent.Agent = (*Agent)(nil)

func loadedContent(loaded skills.Loaded) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Skill: %s\nDescription: %s\n", loaded.Header.Name, loaded.Header.Description)
	if loaded.Reference == "" {
		builder.WriteString("Reference: SKILL.md\n")
	} else {
		fmt.Fprintf(&builder, "Reference: %s\n", loaded.Reference)
	}
	if len(loaded.Header.Metadata) > 0 {
		metadata, err := json.Marshal(loaded.Header.Metadata)
		if err == nil {
			fmt.Fprintf(&builder, "Manifest metadata: %s\n", metadata)
		}
	}
	builder.WriteString("\n")
	builder.WriteString(loaded.Content)
	return builder.String()
}
