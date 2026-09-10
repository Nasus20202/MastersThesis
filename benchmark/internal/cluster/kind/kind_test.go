package kind

import (
	"context"
	"errors"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExecutor struct {
	specs []command.Spec
	err   error
}

func (f *fakeExecutor) Run(_ context.Context, spec command.Spec) (command.Result, error) {
	f.specs = append(f.specs, spec)
	return command.Result{}, f.err
}

func TestNewRequiresExecutorAndName(t *testing.T) {
	_, err := New(nil, Config{Name: "cluster"})
	assert.Error(t, err)

	_, err = New(&fakeExecutor{}, Config{Name: " "})
	assert.Error(t, err)

	_, err = New(&fakeExecutor{}, Config{Name: "cluster", ConfigPath: " "})
	assert.Error(t, err)
}

func TestClusterRunsCreateAndDelete(t *testing.T) {
	executor := &fakeExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark"})
	require.NoError(t, err)

	assert.NoError(t, cluster.Create(context.Background()))
	assert.NoError(t, cluster.Delete(context.Background()))

	require.Len(t, executor.specs, 2)
	kubeconfigPath := cluster.KubeconfigPath()
	assert.Equal(t, program, executor.specs[0].Program)
	assert.Equal(t, []string{"create", "cluster", "--name", "benchmark", "--kubeconfig", kubeconfigPath}, executor.specs[0].Args)
	assert.Equal(t, program, executor.specs[1].Program)
	assert.Equal(t, []string{"delete", "cluster", "--name", "benchmark"}, executor.specs[1].Args)
	assert.Equal(t, "kind-benchmark", cluster.KubeconfigContext())
}

func TestClusterUsesConfigFile(t *testing.T) {
	executor := &fakeExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark", ConfigPath: "/tmp/kind.yaml"})
	require.NoError(t, err)

	assert.NoError(t, cluster.Create(context.Background()))

	require.Len(t, executor.specs, 1)
	kubeconfigPath := cluster.KubeconfigPath()
	assert.Equal(t, program, executor.specs[0].Program)
	assert.Equal(t, "/tmp", executor.specs[0].Dir)
	assert.Equal(t, []string{"create", "cluster", "--name", "benchmark", "--kubeconfig", kubeconfigPath, "--config", "/tmp/kind.yaml"}, executor.specs[0].Args)
}

func TestClusterReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("kind failed")
	cluster, err := New(&fakeExecutor{err: wantErr}, Config{Name: "benchmark"})
	require.NoError(t, err)

	err = cluster.Create(context.Background())
	assert.ErrorIs(t, err, wantErr)
}
