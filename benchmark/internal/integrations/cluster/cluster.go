// Package cluster contains the provider-neutral cluster contract used by the
// benchmark orchestration layer.
package cluster

import "context"

// Cluster is the lifecycle and kubeconfig capability required by the
// benchmark. Provider-specific details remain in adapter packages.
type Cluster interface {
	Create(context.Context) error
	Delete(context.Context) error
	KubeconfigPath() string
	InternalKubeconfigPath() string
	KubeconfigContext() string
}

// Factory constructs a provider-neutral cluster implementation.
type Factory func(string) (Cluster, error)
