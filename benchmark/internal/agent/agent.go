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
