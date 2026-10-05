// Package baseline implements the baseline condition: a bash-only tool loop
// with no system prompt.
package baseline

import (
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const promptInstruction = "You are working in a Kubernetes troubleshooting environment. Use the bash tool to inspect and modify the environment as needed. Complete the task using only the available environment. When finished, provide a short final response."

func New(client inference.Client, shell common.Shell, config common.Config) (*common.Condition, error) {
	tools, err := common.NewSandboxTools(shell, config)
	if err != nil {
		return nil, err
	}
	return common.NewCondition("baseline", client, tools, config, "", baselinePrompt)
}

func baselinePrompt(task string) string {
	return "Task:\n" + task + "\n\n" + promptInstruction
}
