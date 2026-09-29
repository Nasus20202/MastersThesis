// Package sandbox defines the container contracts used by orchestration and
// agent tools.
package sandbox

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
)

// Sandbox is a container that orchestration starts and stops.
type Sandbox interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// Executor runs commands in a started sandbox.
type Executor interface {
	Exec(context.Context, command.Spec) (command.Result, error)
}

// ImageBuilder makes sure the sandbox image exists, building it if needed.
type ImageBuilder interface {
	Build(context.Context) error
}

// Factory creates a sandbox from a cluster name and kubeconfig path.
type Factory func(string, string) (Sandbox, error)
