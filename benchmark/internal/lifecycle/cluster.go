package lifecycle

import "context"

type Cluster interface {
	Create(context.Context) error
	Delete(context.Context) error
	KubeconfigPath() string
	InternalKubeconfigPath() string
	KubeconfigContext() string
}

type ClusterFactory func(string) (Cluster, error)

type Sandbox interface {
	Build(context.Context) error
	Start(context.Context) error
	Stop(context.Context) error
}

type SandboxFactory func(string, string) (Sandbox, error)
