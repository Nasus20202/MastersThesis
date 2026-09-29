// Package skill implements the skill condition: a tool loop that can
// discover and load Kubernetes troubleshooting skills.
package skill

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

//go:embed prompt.md
var defaultSystemPrompt string

//go:embed skills
var defaultSkillFiles embed.FS

// New builds the skill condition. A blank systemPrompt uses the embedded
// prompt.md and nil files use the embedded skills directory. The available
// skills are appended to the system prompt.
func New(client inference.Client, shell common.Shell, config common.Config, systemPrompt string, files fs.FS) (*common.Condition, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = defaultSystemPrompt
	}
	if files == nil {
		files = defaultSkillFiles
	}
	prompt, err := routingSystemPrompt(systemPrompt, files)
	if err != nil {
		return nil, err
	}
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	tools := []common.Tool{bash, skillTool{files: files}, referenceTool{files: files}}
	return common.NewCondition("skill", client, tools, config, prompt, nil)
}
