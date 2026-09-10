package lifecycle

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCluster struct {
	events                 *[]string
	createErr              error
	deleteErr              error
	deleteCtx              context.Context
	kubeconfigPath         string
	internalKubeconfigPath string
	kubeconfigCtx          string
}

func (c *fakeCluster) Create(context.Context) error {
	*c.events = append(*c.events, "create")
	return c.createErr
}

func (c *fakeCluster) Delete(ctx context.Context) error {
	*c.events = append(*c.events, "delete")
	c.deleteCtx = ctx
	return c.deleteErr
}

func (c *fakeCluster) KubeconfigPath() string {
	return c.kubeconfigPath
}

func (c *fakeCluster) InternalKubeconfigPath() string {
	return c.internalKubeconfigPath
}

func (c *fakeCluster) KubeconfigContext() string {
	return c.kubeconfigCtx
}

type recordingExecutor struct {
	events   *[]string
	specs    []command.Spec
	failWith error
	failAt   int
	results  []command.Result
	errors   []error
}

func (e *recordingExecutor) Run(_ context.Context, spec command.Spec) (command.Result, error) {
	e.specs = append(e.specs, spec)
	*e.events = append(*e.events, spec.Program)
	if e.failAt == len(e.specs) {
		return command.Result{}, e.failWith
	}
	resultIndex := len(e.specs) - 1
	if resultIndex >= len(e.results) {
		return command.Result{}, e.errorAt(resultIndex)
	}
	return e.results[resultIndex], e.errorAt(resultIndex)
}

func (e *recordingExecutor) errorAt(index int) error {
	if index >= len(e.errors) {
		return nil
	}
	return e.errors[index]
}

func testDefinition() scenario.Definition {
	return scenario.Definition{
		ID:          "test-scenario",
		Title:       "Test scenario",
		Task:        "Restore the workload.",
		Prepare:     scenario.Step{{Program: "kubectl", Args: []string{"apply", "-f", "manifest.yaml"}}},
		VerifyClean: scenario.Step{{Program: "verify-clean"}},
		InjectFault: scenario.Step{{Program: "inject-fault"}},
		VerifyFault: scenario.Step{{Program: "verify-fault"}},
		Reset:       scenario.Step{{Program: "reset"}},
		Grading: []scenario.Criterion{{
			ID:     "workload-restored",
			Weight: 1,
			Check:  scenario.Command{Program: "verify-restored"},
		}},
	}
}

func TestRunnerRunsPhasesAndCleansUp(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:         &events,
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	executor := &recordingExecutor{events: &events}
	var clusterName string
	runner := Runner{
		ClusterFactory: func(name string) (Cluster, error) {
			clusterName = name
			events = append(events, "factory")
			return cluster, nil
		},
		Executor:       executor,
		CleanupTimeout: time.Second,
	}

	result, err := runner.Run(context.Background(), testDefinition())
	require.NoError(t, err)
	assert.Equal(t, float64(1), result.Grading.Score)
	assert.True(t, result.Grading.FullSuccess)
	assert.Equal(t, "test-scenario", result.ScenarioID)
	assert.True(t, strings.HasPrefix(clusterName, "benchmark-test-scenario-"))

	assert.Equal(t, []string{"factory", "create", "kubectl", "verify-clean", "inject-fault", "verify-fault", "verify-restored", "reset", "delete"}, events)
	require.NotEmpty(t, executor.specs)
	assert.Equal(t, []string{"apply", "-f", "manifest.yaml"}, executor.specs[0].Args)
	assert.Equal(t, "/tmp/test.kubeconfig", executor.specs[0].Env["KUBECONFIG"])
	_, hasDeadline := cluster.deleteCtx.Deadline()
	assert.True(t, hasDeadline)
}

func TestRunGradingRunsCriteriaIndependentlyAndCapturesEvidence(t *testing.T) {
	events := []string{}
	executor := &recordingExecutor{
		events: &events,
		results: []command.Result{
			{Stdout: "healthy", ExitCode: 0, Duration: 10 * time.Millisecond},
			{Stdout: "partial", Stderr: "not ready", ExitCode: 7, Duration: 20 * time.Millisecond},
		},
		errors: []error{nil, errors.New("check failed")},
	}
	runner := Runner{Executor: executor}
	criteria := []scenario.Criterion{
		{ID: "healthy", Weight: 1, Check: scenario.Command{Program: "first-check"}},
		{ID: "ready", Weight: 3, Check: scenario.Command{Program: "second-check"}},
	}

	result, err := runner.runGrading(context.Background(), criteria, "/tmp/test.kubeconfig")
	require.NoError(t, err)
	assert.Equal(t, 0.25, result.Score)
	assert.False(t, result.FullSuccess)
	assert.Equal(t, []string{"first-check", "second-check"}, events)
	require.Len(t, result.Criteria, 2)
	got := result.Criteria[1]
	assert.False(t, got.Passed)
	assert.Equal(t, "check failed", got.Error)
	assert.Equal(t, "partial", got.Stdout)
	assert.Equal(t, "not ready", got.Stderr)
	assert.Equal(t, 7, got.ExitCode)
	assert.Equal(t, 0.02, got.DurationSeconds)
}

func TestWithKubeconfigPreservesEnvironment(t *testing.T) {
	spec := command.Spec{Env: map[string]string{"EXISTING": "value"}}
	got := withKubeconfig(spec, "/tmp/test.kubeconfig")

	assert.Equal(t, "value", got.Env["EXISTING"])
	assert.Equal(t, "/tmp/test.kubeconfig", got.Env[kubeconfigEnv])
	_, modified := spec.Env[kubeconfigEnv]
	assert.False(t, modified)
}

func TestClusterNameUsesScenarioID(t *testing.T) {
	name := clusterNameFor("image-pull-failure")
	assert.True(t, strings.HasPrefix(name, "benchmark-image-pull-failure-"))
}

func TestRunnerCleansUpAfterCreateFailure(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:         &events,
		createErr:      errors.New("create failed"),
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	runner := Runner{
		ClusterFactory: func(string) (Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}

	_, err := runner.Run(context.Background(), testDefinition())
	assert.Error(t, err)
	assert.Equal(t, []string{"create", "delete"}, events)
}

func TestRunnerReturnsCleanupError(t *testing.T) {
	events := []string{}
	cleanupErr := errors.New("delete failed")
	cluster := &fakeCluster{
		events:         &events,
		deleteErr:      cleanupErr,
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	runner := Runner{
		ClusterFactory: func(string) (Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}

	_, err := runner.Run(context.Background(), testDefinition())
	assert.ErrorIs(t, err, cleanupErr)
}
