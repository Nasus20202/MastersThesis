package lifecycle

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
)

type Cluster interface {
	Create(context.Context) error
	Delete(context.Context) error
	Context() string
}

type Runner struct {
	Cluster  Cluster
	Executor command.Executor
}
