// Package prompt implements the prompt condition: a bash-only tool loop
// guided by a system prompt.
package prompt

import (
	_ "embed"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

//go:embed prompt.md
var defaultSystemPrompt string

// New builds the prompt condition. A blank systemPrompt uses the embedded
// prompt.md.
func New(client inference.Client, shell common.Shell, config common.Config, systemPrompt string) (*common.Condition, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = defaultSystemPrompt
	}
	tools, err := common.NewSandboxTools(shell, config)
	if err != nil {
		return nil, err
	}
	return common.NewCondition("prompt", client, tools, config, systemPrompt, nil)
}
