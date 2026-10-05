// Package cluster defines the cluster contract used by orchestration.
package cluster

import "context"

// Cluster is a disposable Kubernetes cluster. InternalKubeconfigPath reaches
// its API server from containers on the kind network.
type Cluster interface {
	Create(context.Context) error
	Delete(context.Context) error
	InternalKubeconfigPath() string
}

// Factory creates a cluster with the given name.
type Factory func(string) (Cluster, error)
