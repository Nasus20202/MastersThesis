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
	"github.com/stretchr/testify/mock"
)

type mockCluster struct {
	mock.Mock
}

func (c *mockCluster) Create(ctx context.Context) error {
	return c.Called(ctx).Error(0)
}

func (c *mockCluster) Delete(ctx context.Context) error {
	return c.Called(ctx).Error(0)
}

func (c *mockCluster) KubeconfigPath() string {
	return c.Called().String(0)
}

func (c *mockCluster) KubeconfigContext() string {
	return c.Called().String(0)
}

type mockExecutor struct {
	mock.Mock
}

func (e *mockExecutor) Run(ctx context.Context, spec command.Spec) (command.Result, error) {
	args := e.Called(ctx, spec)
	result, _ := args.Get(0).(command.Result)
	return result, args.Error(1)
}

func commandSpec(program string, args ...string) interface{} {
	return mock.MatchedBy(func(spec command.Spec) bool {
		if spec.Program != program || len(spec.Args) != len(args) {
			return false
		}
		for index, arg := range args {
			if spec.Args[index] != arg {
				return false
			}
		}
		return spec.Env[kubeconfigEnv] == "/tmp/test.kubeconfig"
	})
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
	cluster := &mockCluster{}
	create := cluster.On("Create", mock.Anything).Return(nil).Once()
	kubeconfigPath := cluster.On("KubeconfigPath").Return("/tmp/test.kubeconfig").Once()
	delete := cluster.On("Delete", mock.MatchedBy(func(ctx context.Context) bool {
		_, ok := ctx.Deadline()
		return ok
	})).Return(nil).Once()
	executor := &mockExecutor{}
	prepare := executor.On("Run", mock.Anything, commandSpec("kubectl", "apply", "-f", "manifest.yaml")).Return(command.Result{}, nil).Once()
	verifyClean := executor.On("Run", mock.Anything, commandSpec("verify-clean")).Return(command.Result{}, nil).Once()
	injectFault := executor.On("Run", mock.Anything, commandSpec("inject-fault")).Return(command.Result{}, nil).Once()
	verifyFault := executor.On("Run", mock.Anything, commandSpec("verify-fault")).Return(command.Result{}, nil).Once()
	grade := executor.On("Run", mock.Anything, commandSpec("verify-restored")).Return(command.Result{}, nil).Once()
	reset := executor.On("Run", mock.Anything, commandSpec("reset")).Return(command.Result{}, nil).Once()
	mock.InOrder(create, kubeconfigPath, prepare, verifyClean, injectFault, verifyFault, grade, reset, delete)
	var clusterName string
	runner := Runner{
		ClusterFactory: func(name string) (Cluster, error) {
			clusterName = name
			return cluster, nil
		},
		Executor:       executor,
		CleanupTimeout: time.Second,
	}

	result, err := runner.Run(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Grading.Score != 1 || !result.Grading.FullSuccess {
		t.Fatalf("grading result = %#v, want full success with score 1", result)
	}
	if result.ScenarioID != "test-scenario" {
		t.Fatalf("scenario ID = %q, want test-scenario", result.ScenarioID)
	}
	if !strings.HasPrefix(clusterName, "benchmark-test-scenario-") {
		t.Fatalf("cluster name = %q, want scenario prefix", clusterName)
	}
	assert.True(t, cluster.AssertExpectations(t))
	assert.True(t, executor.AssertExpectations(t))
}

func TestRunGradingRunsCriteriaIndependentlyAndCapturesEvidence(t *testing.T) {
	executor := &mockExecutor{}
	first := executor.On("Run", mock.Anything, commandSpec("first-check")).Return(command.Result{
		Stdout:   "healthy",
		ExitCode: 0,
		Duration: 10 * time.Millisecond,
	}, nil).Once()
	second := executor.On("Run", mock.Anything, commandSpec("second-check")).Return(command.Result{
		Stdout:   "partial",
		Stderr:   "not ready",
		ExitCode: 7,
		Duration: 20 * time.Millisecond,
	}, errors.New("check failed")).Once()
	mock.InOrder(first, second)
	runner := Runner{Executor: executor}
	criteria := []scenario.Criterion{
		{ID: "healthy", Weight: 1, Check: scenario.Command{Program: "first-check"}},
		{ID: "ready", Weight: 3, Check: scenario.Command{Program: "second-check"}},
	}

	result, err := runner.runGrading(context.Background(), criteria, "/tmp/test.kubeconfig")
	if err != nil {
		t.Fatalf("run grading: %v", err)
	}
	if result.Score != 0.25 || result.FullSuccess {
		t.Fatalf("grading result = %#v, want score 0.25 and partial success", result)
	}
	if len(result.Criteria) != 2 {
		t.Fatalf("criteria = %#v, want two results", result.Criteria)
	}
	if got := result.Criteria[1]; got.Passed || got.Error != "check failed" || got.Stdout != "partial" || got.Stderr != "not ready" || got.ExitCode != 7 || got.DurationSeconds != 0.02 {
		t.Fatalf("failed criterion evidence = %#v, want captured result", got)
	}
	assert.True(t, executor.AssertExpectations(t))
}

func TestWithKubeconfigPreservesEnvironment(t *testing.T) {
	spec := command.Spec{Env: map[string]string{"EXISTING": "value"}}
	got := withKubeconfig(spec, "/tmp/test.kubeconfig")

	if got.Env["EXISTING"] != "value" {
		t.Fatalf("EXISTING = %q, want value", got.Env["EXISTING"])
	}
	if got.Env[kubeconfigEnv] != "/tmp/test.kubeconfig" {
		t.Fatalf("KUBECONFIG = %q, want %q", got.Env[kubeconfigEnv], "/tmp/test.kubeconfig")
	}
	if _, ok := spec.Env[kubeconfigEnv]; ok {
		t.Fatal("withKubeconfig modified the original environment")
	}
}

func TestClusterNameUsesScenarioID(t *testing.T) {
	name := clusterNameFor("image-pull-failure")
	if !strings.HasPrefix(name, "benchmark-image-pull-failure-") {
		t.Fatalf("cluster name = %q, want scenario prefix", name)
	}
}

func TestRunnerCleansUpAfterCreateFailure(t *testing.T) {
	createErr := errors.New("create failed")
	cluster := &mockCluster{}
	create := cluster.On("Create", mock.Anything).Return(createErr).Once()
	delete := cluster.On("Delete", mock.Anything).Return(nil).Once()
	mock.InOrder(create, delete)
	runner := Runner{
		ClusterFactory: func(string) (Cluster, error) { return cluster, nil },
		Executor:       &mockExecutor{},
	}

	if _, err := runner.Run(context.Background(), testDefinition()); err == nil {
		t.Fatal("expected create error")
	}
	assert.True(t, cluster.AssertExpectations(t))
}

func TestRunnerReturnsCleanupError(t *testing.T) {
	cleanupErr := errors.New("delete failed")
	cluster := &mockCluster{}
	create := cluster.On("Create", mock.Anything).Return(nil).Once()
	kubeconfigPath := cluster.On("KubeconfigPath").Return("/tmp/test.kubeconfig").Once()
	delete := cluster.On("Delete", mock.Anything).Return(cleanupErr).Once()
	executor := &mockExecutor{}
	prepare := executor.On("Run", mock.Anything, commandSpec("kubectl", "apply", "-f", "manifest.yaml")).Return(command.Result{}, nil).Once()
	verifyClean := executor.On("Run", mock.Anything, commandSpec("verify-clean")).Return(command.Result{}, nil).Once()
	injectFault := executor.On("Run", mock.Anything, commandSpec("inject-fault")).Return(command.Result{}, nil).Once()
	verifyFault := executor.On("Run", mock.Anything, commandSpec("verify-fault")).Return(command.Result{}, nil).Once()
	grade := executor.On("Run", mock.Anything, commandSpec("verify-restored")).Return(command.Result{}, nil).Once()
	reset := executor.On("Run", mock.Anything, commandSpec("reset")).Return(command.Result{}, nil).Once()
	mock.InOrder(create, kubeconfigPath, prepare, verifyClean, injectFault, verifyFault, grade, reset, delete)
	runner := Runner{
		ClusterFactory: func(string) (Cluster, error) { return cluster, nil },
		Executor:       executor,
	}

	_, err := runner.Run(context.Background(), testDefinition())
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("Run() error = %v, want cleanup error", err)
	}
	assert.True(t, cluster.AssertExpectations(t))
	assert.True(t, executor.AssertExpectations(t))
}
