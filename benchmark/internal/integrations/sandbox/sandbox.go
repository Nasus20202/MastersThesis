// Package sandbox contains the provider-neutral sandbox contracts used by the
// benchmark orchestration and agent wiring layers.
package sandbox

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
)

// Sandbox is the lifecycle capability required by orchestration.
type Sandbox interface {
	Build(context.Context) error
	Start(context.Context) error
	Stop(context.Context) error
}

// Executor is the command-execution capability exposed by a running sandbox.
// It is deliberately separate from Sandbox lifecycle management.
type Executor interface {
	Exec(context.Context, command.Spec) (command.Result, error)
}

// ImageBuilder builds or ensures the sandbox image is available.
type ImageBuilder interface {
	Build(context.Context) error
}

// Factory constructs a provider-neutral sandbox implementation.
type Factory func(string, string) (Sandbox, error)
