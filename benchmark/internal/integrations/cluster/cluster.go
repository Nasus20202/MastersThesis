// Package cluster defines the cluster contract used by orchestration.
package cluster

import "context"

// Cluster is a disposable Kubernetes cluster and its kubeconfig.
type Cluster interface {
	Create(context.Context) error
	Delete(context.Context) error
	KubeconfigPath() string
	InternalKubeconfigPath() string
	KubeconfigContext() string
}

// Factory creates a cluster with the given name.
type Factory func(string) (Cluster, error)
