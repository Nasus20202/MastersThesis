package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/skills"
)

const loadSkillToolName = "load_skill"

type loadSkillTool struct {
	registry *skills.Registry
}

// SkillLoadEvidence records which allowlisted skill content was returned.
type SkillLoadEvidence struct {
	Skill     string `json:"skill"`
	Reference string `json:"reference"`
	Bytes     int    `json:"bytes"`
}

func NewLoadSkillTool(registry *skills.Registry) (common.Tool, error) {
	if registry == nil {
		return nil, errors.New("skills registry is required")
	}
	return loadSkillTool{registry: registry}, nil
}

func (loadSkillTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        loadSkillToolName,
		Description: "Load an approved skill or one of its listed reference files by name.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"skill":{"type":"string","description":"Skill name from the manifest catalog"},"reference":{"type":"string","description":"Optional reference filename listed for the skill"}},"required":["skill"],"additionalProperties":false}`),
	}
}

func (t loadSkillTool) Execute(ctx context.Context, call inference.ToolCall) common.ToolResult {
	if err := ctx.Err(); err != nil {
		return common.ToolResult{Content: err.Error(), Error: err}
	}
	var arguments struct {
		Skill     string `json:"skill"`
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		wrapped := fmt.Errorf("malformed load_skill arguments: %w", err)
		return common.ToolResult{Content: wrapped.Error(), Error: wrapped}
	}
	arguments.Skill = strings.TrimSpace(arguments.Skill)
	arguments.Reference = strings.TrimSpace(arguments.Reference)
	loaded, err := t.registry.Load(arguments.Skill, arguments.Reference)
	if err != nil {
		wrapped := fmt.Errorf("load skill: %w", err)
		return common.ToolResult{Content: wrapped.Error(), Error: wrapped}
	}
	reference := loaded.Reference
	if reference == "" {
		reference = skills.MainFilename
	}
	return common.ToolResult{
		Content: loadedContent(loaded),
		Details: SkillLoadEvidence{
			Skill:     loaded.Header.Name,
			Reference: reference,
			Bytes:     len(loaded.Content),
		},
	}
}
