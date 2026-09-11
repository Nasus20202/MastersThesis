package orchestration

import "context"

type Sandbox interface {
	Build(context.Context) error
	Start(context.Context) error
	Stop(context.Context) error
}

type SandboxImageBuilder interface {
	Build(context.Context) error
}

type SandboxFactory func(string, string) (Sandbox, error)
