package agent

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
)

// Agent is the model-driven boundary used by benchmark conditions.
type Agent interface {
	Run(context.Context, string) (common.Result, error)
}
