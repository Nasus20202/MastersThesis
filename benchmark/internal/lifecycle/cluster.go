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
