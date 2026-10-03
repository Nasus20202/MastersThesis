// Package skill implements the skill condition: a tool loop that can
// discover and load Kubernetes troubleshooting skills.
package skill

import (
	"context"
	"embed"
	"fmt"
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
	var hint common.Hint
	if section := changeSection(files); section != "" {
		hint = (&changeHint{section: section}).hint
	}
	bash, err := common.NewHintedBashTool(shell, hint)
	if err != nil {
		return nil, err
	}
	tools := []common.Tool{bash, skillTool{files: files}, referenceTool{files: files}}
	return common.NewCondition("skill", client, tools, config, prompt, nil)
}

// The change hint: when a kubectl change fails, the harness appends the
// kubectl skill's section on changing state to the Bash result once per
// attempt, so the curated mechanics are in view at the failure.
const (
	changeSkill   = "kubectl"
	changeHeading = "## Changing state"
)

// ChangeHintEvidence records the appended section for later analysis.
type ChangeHintEvidence struct {
	Skill   string `json:"skill"`
	Section string `json:"section"`
}

// changeSection is the section under changeHeading in the kubectl skill, or
// empty when the skill set has no such section.
func changeSection(files fs.FS) string {
	_, body, err := readSkillDocument(files, changeSkill)
	if err != nil {
		return ""
	}
	_, section, found := strings.Cut(body, "\n"+changeHeading+"\n")
	if !found {
		return ""
	}
	if end := strings.Index(section, "\n## "); end >= 0 {
		section = section[:end]
	}
	return strings.TrimSpace(section)
}

// changeHint holds one attempt's state; New builds one per attempt.
type changeHint struct {
	section string
	shown   bool
}

func (h *changeHint) hint(context.Context, common.FailedChange) (string, any) {
	if h.shown {
		return "", nil
	}
	h.shown = true
	return fmt.Sprintf("Guidance from the %s skill on changing state:\n%s", changeSkill, h.section),
		ChangeHintEvidence{Skill: changeSkill, Section: strings.TrimPrefix(changeHeading, "## ")}
}
