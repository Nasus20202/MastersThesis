package kind

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExecutor struct {
	specs   []command.Spec
	results []command.Result
	err     error
}

func (f *fakeExecutor) Run(_ context.Context, spec command.Spec) (command.Result, error) {
	f.specs = append(f.specs, spec)
	resultIndex := len(f.specs) - 1
	if resultIndex >= len(f.results) {
		return command.Result{}, f.err
	}
	return f.results[resultIndex], f.err
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
	executor := &fakeExecutor{results: []command.Result{{}, {Stdout: "apiVersion: v1\n"}, {}}}
	cluster, err := New(executor, Config{Name: "benchmark"})
	require.NoError(t, err)

	assert.NoError(t, cluster.Create(context.Background()))
	assert.NoError(t, cluster.Delete(context.Background()))

	require.Len(t, executor.specs, 3)
	kubeconfigPath := cluster.KubeconfigPath()
	assert.Equal(t, program, executor.specs[0].Program)
	assert.Equal(t, []string{"create", "cluster", "--name", "benchmark", "--kubeconfig", kubeconfigPath}, executor.specs[0].Args)
	assert.Equal(t, program, executor.specs[1].Program)
	assert.Equal(t, []string{"get", "kubeconfig", "--name", "benchmark", "--internal"}, executor.specs[1].Args)
	assert.Equal(t, program, executor.specs[2].Program)
	assert.Equal(t, []string{"delete", "cluster", "--name", "benchmark"}, executor.specs[2].Args)
	assert.Equal(t, "kind-benchmark", cluster.KubeconfigContext())
	assert.NotEqual(t, kubeconfigPath, cluster.InternalKubeconfigPath())
}

func TestClusterCreateGeneratesInternalKubeconfig(t *testing.T) {
	executor := &fakeExecutor{results: []command.Result{{}, {Stdout: "apiVersion: v1\n"}}}
	cluster, err := New(executor, Config{Name: "benchmark-internal"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(cluster.InternalKubeconfigPath())) })

	assert.NoError(t, cluster.Create(context.Background()))

	require.Len(t, executor.specs, 2)
	assert.Equal(t, command.Spec{
		Program: program,
		Args:    []string{"get", "kubeconfig", "--name", "benchmark-internal", "--internal"},
	}, executor.specs[1])
	contents, err := os.ReadFile(cluster.InternalKubeconfigPath())
	require.NoError(t, err)
	assert.Equal(t, "apiVersion: v1\n", string(contents))
	info, err := os.Stat(cluster.InternalKubeconfigPath())
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestClusterRejectsEmptyInternalKubeconfig(t *testing.T) {
	cluster, err := New(&fakeExecutor{results: []command.Result{{}, {}}}, Config{Name: "benchmark-empty"})
	require.NoError(t, err)

	err = cluster.Create(context.Background())
	assert.EqualError(t, err, "get internal kubeconfig for cluster \"benchmark-empty\": command returned empty output")
}

func TestClusterUsesConfigFile(t *testing.T) {
	executor := &fakeExecutor{results: []command.Result{{}, {Stdout: "apiVersion: v1\n"}}}
	cluster, err := New(executor, Config{Name: "benchmark", ConfigPath: "/tmp/kind.yaml"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(cluster.InternalKubeconfigPath())) })

	assert.NoError(t, cluster.Create(context.Background()))

	require.Len(t, executor.specs, 2)
	kubeconfigPath := cluster.KubeconfigPath()
	assert.Equal(t, program, executor.specs[0].Program)
	assert.Equal(t, "/tmp", executor.specs[0].Dir)
	assert.Equal(t, []string{"create", "cluster", "--name", "benchmark", "--kubeconfig", kubeconfigPath, "--config", "/tmp/kind.yaml"}, executor.specs[0].Args)
	assert.Equal(t, []string{"get", "kubeconfig", "--name", "benchmark", "--internal"}, executor.specs[1].Args)
}

func TestClusterReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("kind failed")
	cluster, err := New(&fakeExecutor{err: wantErr}, Config{Name: "benchmark"})
	require.NoError(t, err)

	err = cluster.Create(context.Background())
	assert.ErrorIs(t, err, wantErr)
}

func TestClusterIncludesCommandStderrInExecutorError(t *testing.T) {
	wantErr := errors.New("kind failed")
	executor := &fakeExecutor{err: wantErr, results: []command.Result{{Stderr: "node name is too long\n"}}}
	cluster, err := New(executor, Config{Name: "benchmark"})
	require.NoError(t, err)

	err = cluster.Create(context.Background())
	assert.ErrorContains(t, err, "node name is too long")
}
