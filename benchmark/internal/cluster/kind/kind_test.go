package kind

import (
	"context"
	"errors"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockExecutor struct {
	mock.Mock
}

func (e *mockExecutor) Run(ctx context.Context, spec command.Spec) (command.Result, error) {
	args := e.Called(ctx, spec)
	result, _ := args.Get(0).(command.Result)
	return result, args.Error(1)
}

func TestNewRequiresExecutorAndName(t *testing.T) {
	_, err := New(nil, Config{Name: "cluster"})
	assert.Error(t, err)

	_, err = New(&mockExecutor{}, Config{Name: " "})
	assert.Error(t, err)

	_, err = New(&mockExecutor{}, Config{Name: "cluster", ConfigPath: " "})
	assert.Error(t, err)
}

func TestClusterRunsCreateAndDelete(t *testing.T) {
	executor := &mockExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark"})
	if !assert.NoError(t, err) {
		return
	}

	kubeconfigPath := cluster.KubeconfigPath()
	create := executor.On("Run", mock.Anything, command.Spec{
		Program: program,
		Args:    []string{"create", "cluster", "--name", "benchmark", "--kubeconfig", kubeconfigPath},
	}).Return(command.Result{}, nil).Once()
	delete := executor.On("Run", mock.Anything, command.Spec{
		Program: program,
		Args:    []string{"delete", "cluster", "--name", "benchmark"},
	}).Return(command.Result{}, nil).Once()
	mock.InOrder(create, delete)

	assert.NoError(t, cluster.Create(context.Background()))
	assert.NoError(t, cluster.Delete(context.Background()))
	assert.Equal(t, "kind-benchmark", cluster.KubeconfigContext())
	assert.True(t, executor.AssertExpectations(t))
}

func TestClusterUsesConfigFile(t *testing.T) {
	executor := &mockExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark", ConfigPath: "/tmp/kind.yaml"})
	if !assert.NoError(t, err) {
		return
	}

	executor.On("Run", mock.Anything, command.Spec{
		Program: program,
		Args:    []string{"create", "cluster", "--name", "benchmark", "--kubeconfig", cluster.KubeconfigPath(), "--config", "/tmp/kind.yaml"},
		Dir:     "/tmp",
	}).Return(command.Result{}, nil).Once()

	assert.NoError(t, cluster.Create(context.Background()))
	assert.True(t, executor.AssertExpectations(t))
}

func TestClusterReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("kind failed")
	executor := &mockExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark"})
	if !assert.NoError(t, err) {
		return
	}

	executor.On("Run", mock.Anything, mock.Anything).Return(command.Result{}, wantErr).Once()
	err = cluster.Create(context.Background())
	assert.ErrorIs(t, err, wantErr)
	assert.True(t, executor.AssertExpectations(t))
}
