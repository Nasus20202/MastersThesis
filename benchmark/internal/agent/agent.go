// Package agent defines the Agent and Factory interfaces implemented by the
// benchmark conditions.
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

// Continuer is implemented by agents that can continue a recorded conversation
// in a fresh cluster (the S2 training variant, #25).
type Continuer interface {
	Continue(context.Context, string, common.Prefix) (common.Result, error)
}

// Factory constructs an agent with the executor for the current sandbox run.
type Factory func(sandboxintegration.Executor) (Agent, error)
