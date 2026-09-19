package agent

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
)

// Agent is the model-driven boundary used by benchmark conditions.
type Agent interface {
	Run(context.Context, string) (common.Result, error)
}

// Factory constructs an agent with the executor for the current sandbox run.
type Factory func(sandboxintegration.Executor) (Agent, error)

// ExecutionContract describes the operational requirements shared by hands-on
// benchmark conditions.
const ExecutionContract = `Use Bash to inspect the sandbox directly and discover any needed resource names, namespaces, cluster state, events, logs, and other details with Bash/kubectl. Diagnose the issue, apply the required repair, and verify the resulting state. Do not ask the user for information available from the sandbox or stop after explaining a plan or suggesting commands. Continue until the task is repaired and verified, or no further progress can reasonably be made.`

// WithExecutionContract prepends the shared operational requirements to a condition prompt.
func WithExecutionContract(prompt string) string {
	return ExecutionContract + "\n\n" + prompt
}
