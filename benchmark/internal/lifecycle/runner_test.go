package lifecycle

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

type fakeCluster struct {
	events         *[]string
	createErr      error
	deleteErr      error
	deleteCtx      context.Context
	kubeconfigPath string
	kubeconfigCtx  string
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

func (c *fakeCluster) KubeconfigContext() string {
	return c.kubeconfigCtx
}

type recordingExecutor struct {
	events   *[]string
	specs    []command.Spec
	failWith error
	failAt   int
}

func (e *recordingExecutor) Run(_ context.Context, spec command.Spec) (command.Result, error) {
	e.specs = append(e.specs, spec)
	*e.events = append(*e.events, spec.Program)
	if e.failAt == len(e.specs) {
		return command.Result{}, e.failWith
	}
	return command.Result{}, nil
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

	if err := runner.Run(context.Background(), testDefinition()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.HasPrefix(clusterName, "benchmark-test-scenario-") {
		t.Fatalf("cluster name = %q, want scenario prefix", clusterName)
	}

	wantEvents := []string{"factory", "create", "kubectl", "verify-clean", "inject-fault", "verify-fault", "reset", "delete"}
	if !slices.Equal(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if got := executor.specs[0].Args; !slices.Equal(got, []string{"apply", "-f", "manifest.yaml"}) {
		t.Fatalf("kubectl args = %v, want unchanged args", got)
	}
	if got := executor.specs[0].Env["KUBECONFIG"]; got != "/tmp/test.kubeconfig" {
		t.Fatalf("KUBECONFIG = %q, want %q", got, "/tmp/test.kubeconfig")
	}
	if _, ok := cluster.deleteCtx.Deadline(); !ok {
		t.Fatal("cleanup context has no deadline")
	}
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

	if err := runner.Run(context.Background(), testDefinition()); err == nil {
		t.Fatal("expected create error")
	}
	if len(events) != 2 || events[0] != "create" || events[1] != "delete" {
		t.Fatalf("events = %v, want create and delete", events)
	}
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

	err := runner.Run(context.Background(), testDefinition())
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("Run() error = %v, want cleanup error", err)
	}
}
