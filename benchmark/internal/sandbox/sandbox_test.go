package sandbox

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

func newSandbox(t *testing.T, executor command.Executor) *Sandbox {
	t.Helper()
	sandbox, err := New(executor, Config{
		Name:           "benchmark-sandbox",
		Image:          "masters-thesis-sandbox:increment-1",
		DockerfilePath: "/tmp/Dockerfile",
		BuildContext:   "/tmp/context",
		KubeconfigPath: "/tmp/benchmark.kubeconfig",
		Network:        "benchmark-sandbox-network",
		NetworkTarget:  "benchmark-control-plane",
		Layout:         DefaultImageLayout(),
	})
	require.NoError(t, err)
	return sandbox
}

func TestNewValidatesConfiguration(t *testing.T) {
	_, err := New(nil, Config{})
	assert.Error(t, err)

	executor := &fakeExecutor{}
	_, err = New(executor, Config{Image: "image", KubeconfigPath: "/tmp/config"})
	assert.Error(t, err)
	_, err = New(executor, Config{Name: "name", KubeconfigPath: "/tmp/config"})
	assert.Error(t, err)
	_, err = New(executor, Config{Name: "name", Image: "image", KubeconfigPath: "/tmp/config"})
	assert.Error(t, err)
	_, err = New(executor, Config{Name: "name", Image: "image", DockerfilePath: "/tmp/Dockerfile", KubeconfigPath: "/tmp/config"})
	assert.Error(t, err)
	_, err = New(executor, Config{Name: "name", Image: "image", BuildContext: "/tmp/context", KubeconfigPath: "/tmp/config"})
	assert.Error(t, err)
}

func TestBuildUsesHostDockerCommand(t *testing.T) {
	executor := &fakeExecutor{}
	sandbox := newSandbox(t, executor)

	assert.NoError(t, sandbox.Build(context.Background()))

	require.Len(t, executor.specs, 1)
	assert.Equal(t, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"build", "--file", "/tmp/Dockerfile",
			"--tag", "masters-thesis-sandbox:increment-1", "/tmp/context",
		},
	}, executor.specs[0])
}

func TestStartUsesRestrictedHostDockerCommand(t *testing.T) {
	executor := &fakeExecutor{}
	sandbox := newSandbox(t, executor)

	assert.NoError(t, sandbox.Start(context.Background()))

	require.Len(t, executor.specs, 4)
	assert.Equal(t, command.Spec{
		Program: dockerProgram,
		Args:    []string{"network", "create", "--internal", "benchmark-sandbox-network"},
	}, executor.specs[0])
	assert.Equal(t, command.Spec{
		Program: dockerProgram,
		Args:    []string{"network", "connect", "benchmark-sandbox-network", "benchmark-control-plane"},
	}, executor.specs[1])
	assert.Equal(t, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"run", "--detach", "--rm", "--name", "benchmark-sandbox",
			"--network", "benchmark-sandbox-network",
			"--user", "benchmark",
			"--workdir", "/workspace",
			"--cap-drop", "ALL",
			"--security-opt", "no-new-privileges",
			"--read-only",
			"--tmpfs", "/tmp:rw,mode=1777",
			"--tmpfs", "/workspace:rw,mode=1777",
			"--mount", "type=bind,src=/tmp/benchmark.kubeconfig,dst=/home/benchmark/.kube/config,readonly",
			"masters-thesis-sandbox:increment-1", "sleep", "infinity",
		},
	}, executor.specs[2])
	assert.Equal(t, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"exec", "--user", "benchmark", "--workdir", "/workspace",
			"benchmark-sandbox", "kubectl", "get", "nodes",
		},
	}, executor.specs[3])
	assert.NotContains(t, executor.specs[2].Args, "/var/run/docker.sock")
}

func TestExecUsesDockerExecWithStableEnvironmentAndWorkdir(t *testing.T) {
	executor := &fakeExecutor{}
	sandbox := newSandbox(t, executor)

	result, err := sandbox.Exec(context.Background(), command.Spec{
		Program: "bash",
		Args:    []string{"-lc", "printf '%s' \"$MODE\""},
		Dir:     "/workspace/task",
		Env: map[string]string{
			"ZED":  "last",
			"MODE": "raw",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, command.Result{}, result)
	require.Len(t, executor.specs, 1)
	assert.Equal(t, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"exec", "--user", "benchmark", "--workdir", "/workspace/task",
			"--env", "MODE=raw", "--env", "ZED=last",
			"benchmark-sandbox", "bash", "-lc", "printf '%s' \"$MODE\"",
		},
	}, executor.specs[0])
}

func TestExecDefaultsToWorkspace(t *testing.T) {
	executor := &fakeExecutor{}
	sandbox := newSandbox(t, executor)

	_, err := sandbox.Exec(context.Background(), command.Spec{Program: "pwd"})
	assert.NoError(t, err)

	require.Len(t, executor.specs, 1)
	assert.Equal(t, "/workspace", executor.specs[0].Args[4])
}

func TestStopUsesHostDockerCommand(t *testing.T) {
	executor := &fakeExecutor{}
	sandbox := newSandbox(t, executor)

	assert.NoError(t, sandbox.Stop(context.Background()))

	require.Len(t, executor.specs, 3)
	assert.Equal(t, command.Spec{Program: dockerProgram, Args: []string{"stop", "benchmark-sandbox"}}, executor.specs[0])
	assert.Equal(t, command.Spec{Program: dockerProgram, Args: []string{"network", "disconnect", "--force", "benchmark-sandbox-network", "benchmark-control-plane"}}, executor.specs[1])
	assert.Equal(t, command.Spec{Program: dockerProgram, Args: []string{"network", "rm", "benchmark-sandbox-network"}}, executor.specs[2])
}

func TestSandboxReturnsDockerErrors(t *testing.T) {
	wantErr := errors.New("docker failed")
	sandbox := newSandbox(t, &fakeExecutor{err: wantErr})

	err := sandbox.Start(context.Background())
	assert.ErrorIs(t, err, wantErr)

	err = sandbox.Build(context.Background())
	assert.ErrorIs(t, err, wantErr)

	_, err = sandbox.Exec(context.Background(), command.Spec{Program: "pwd"})
	assert.ErrorIs(t, err, wantErr)

	err = sandbox.Stop(context.Background())
	assert.ErrorIs(t, err, wantErr)
}
