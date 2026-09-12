package orchestration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	clusterintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDefinition() scenario.Definition {
	return scenario.Definition{
		ID:          "test-scenario",
		Title:       "Test scenario",
		Task:        "Restore the workload.",
		Prepare:     scenario.Step{{Program: "kubectl", Args: []string{"apply", "-f", "manifest.yaml"}}},
		VerifyClean: scenario.Step{{Program: "verify-clean"}},
		InjectFault: scenario.Step{{Program: "inject-fault"}},
		VerifyFault: scenario.Step{{Program: "verify-fault"}},
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
		ClusterFactory: func(name string) (clusterintegration.Cluster, error) {
			clusterName = name
			events = append(events, "factory")
			return cluster, nil
		},
		Executor:       executor,
		CleanupTimeout: time.Second,
	}

	result, err := runner.RunWithRepair(context.Background(), testDefinition(), nil)
	require.NoError(t, err)
	assert.Equal(t, float64(1), result.Grading.Score)
	assert.True(t, result.Grading.FullSuccess)
	assert.Equal(t, "test-scenario", result.ScenarioID)
	assert.True(t, strings.HasPrefix(clusterName, "benchmark-test-scenario-"))

	assert.Equal(t, []string{"factory", "create", "kubectl", "verify-clean", "inject-fault", "verify-fault", "verify-restored", "delete"}, events)
	require.NotEmpty(t, executor.specs)
	assert.Equal(t, []string{"apply", "-f", "manifest.yaml"}, executor.specs[0].Args)
	assert.Equal(t, "/tmp/test.kubeconfig", executor.specs[0].Env["KUBECONFIG"])
	_, hasDeadline := cluster.deleteCtx.Deadline()
	assert.True(t, hasDeadline)
}

func TestRunnerBuildsStartsAndStopsSandboxAroundPhases(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:                 &events,
		kubeconfigPath:         "/tmp/test.kubeconfig",
		internalKubeconfigPath: "/tmp/test.internal.kubeconfig",
		kubeconfigCtx:          "kind-test",
	}
	sandbox := &fakeSandbox{events: &events}
	executor := &recordingExecutor{events: &events}
	runner := Runner{
		ClusterFactory: func(name string) (clusterintegration.Cluster, error) {
			events = append(events, "factory")
			return cluster, nil
		},
		SandboxFactory: func(name, kubeconfigPath string) (sandboxintegration.Sandbox, error) {
			assert.True(t, strings.HasPrefix(name, "benchmark-test-scenario-"))
			assert.Equal(t, "/tmp/test.internal.kubeconfig", kubeconfigPath)
			events = append(events, "sandbox-factory")
			return sandbox, nil
		},
		Executor:       executor,
		CleanupTimeout: time.Second,
	}

	_, err := runner.RunWithRepair(context.Background(), testDefinition(), nil)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"factory", "create", "sandbox-factory", "sandbox-build", "sandbox-start",
		"kubectl", "verify-clean", "inject-fault", "verify-fault", "verify-restored",
		"sandbox-stop", "delete",
	}, events)
}

func TestRunnerUsesSharedSandboxImageBuilder(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:                 &events,
		kubeconfigPath:         "/tmp/test.kubeconfig",
		internalKubeconfigPath: "/tmp/test.internal.kubeconfig",
		kubeconfigCtx:          "kind-test",
	}
	sandbox := &fakeSandbox{events: &events}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) {
			events = append(events, "factory")
			return cluster, nil
		},
		SandboxFactory: func(string, string) (sandboxintegration.Sandbox, error) {
			events = append(events, "sandbox-factory")
			return sandbox, nil
		},
		SandboxImageBuilder: &fakeImageBuilder{events: &events},
		Executor:            &recordingExecutor{events: &events},
		CleanupTimeout:      time.Second,
	}

	_, err := runner.RunWithRepair(context.Background(), testDefinition(), nil)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"sandbox-image-build", "factory", "create", "sandbox-factory", "sandbox-start",
		"kubectl", "verify-clean", "inject-fault", "verify-fault", "verify-restored",
		"sandbox-stop", "delete",
	}, events)
}

func TestRunnerRunsInjectedAgentAfterFaultVerification(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:                 &events,
		kubeconfigPath:         "/tmp/test.kubeconfig",
		internalKubeconfigPath: "/tmp/test.internal.kubeconfig",
		kubeconfigCtx:          "kind-test",
	}
	sandbox := &fakeSandbox{events: &events}
	agent := &fakeAgent{events: &events}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		SandboxFactory: func(string, string) (sandboxintegration.Sandbox, error) { return sandbox, nil },
		AgentFactory: rootagent.Factory(func(executor sandboxintegration.Executor) (rootagent.Agent, error) {
			assert.Same(t, sandbox, executor)
			events = append(events, "agent-factory")
			return agent, nil
		}),
		Executor: &recordingExecutor{events: &events},
	}

	result, err := runner.Run(context.Background(), testDefinition())
	require.NoError(t, err)
	assert.Equal(t, testDefinition().Task, agent.task)
	require.NotNil(t, result.Agent)
	assert.Equal(t, testDefinition().Task, result.Agent.Task)
	assert.Equal(t, []string{
		"create", "sandbox-build", "sandbox-start", "agent-factory", "kubectl", "verify-clean",
		"inject-fault", "verify-fault", "agent", "verify-restored", "sandbox-stop", "delete",
	}, events)
}

func TestRunnerPreservesAgentFailureAndStillCleansUp(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:                 &events,
		kubeconfigPath:         "/tmp/test.kubeconfig",
		internalKubeconfigPath: "/tmp/test.internal.kubeconfig",
	}
	sandbox := &fakeSandbox{events: &events}
	wantErr := errors.New("model stopped unexpectedly")
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		SandboxFactory: func(string, string) (sandboxintegration.Sandbox, error) { return sandbox, nil },
		AgentFactory: rootagent.Factory(func(sandboxintegration.Executor) (rootagent.Agent, error) {
			return &fakeAgent{events: &events, err: wantErr}, nil
		}),
		Executor: &recordingExecutor{events: &events},
	}

	result, err := runner.Run(context.Background(), testDefinition())
	assert.ErrorIs(t, err, wantErr)
	require.NotNil(t, result.Agent)
	assert.Equal(t, common.TerminationCompleted, result.Agent.Termination)
	assert.Equal(t, []string{
		"create", "sandbox-build", "sandbox-start", "kubectl", "verify-clean", "inject-fault", "verify-fault",
		"agent", "verify-restored", "sandbox-stop", "delete",
	}, events)
}

func TestRunnerRequiresSandboxExecutorForAgent(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:                 &events,
		kubeconfigPath:         "/tmp/test.kubeconfig",
		internalKubeconfigPath: "/tmp/test.internal.kubeconfig",
	}
	sandbox := &lifecycleOnlySandbox{events: &events}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		SandboxFactory: func(string, string) (sandboxintegration.Sandbox, error) { return sandbox, nil },
		AgentFactory: rootagent.Factory(func(sandboxintegration.Executor) (rootagent.Agent, error) {
			return &fakeAgent{}, nil
		}),
		Executor: &recordingExecutor{events: &events},
	}

	_, err := runner.Run(context.Background(), testDefinition())
	assert.ErrorContains(t, err, "does not provide command execution")
	assert.Equal(t, []string{"create", "sandbox-build", "sandbox-start", "sandbox-stop", "delete"}, events)
}

func TestRunnerRejectsInvalidAgentFactoryResults(t *testing.T) {
	for name, factory := range map[string]rootagent.Factory{
		"error": func(sandboxintegration.Executor) (rootagent.Agent, error) {
			return nil, errors.New("factory failed")
		},
		"nil agent": func(sandboxintegration.Executor) (rootagent.Agent, error) {
			return nil, nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			events := []string{}
			cluster := &fakeCluster{
				events:                 &events,
				kubeconfigPath:         "/tmp/test.kubeconfig",
				internalKubeconfigPath: "/tmp/test.internal.kubeconfig",
			}
			sandbox := &fakeSandbox{events: &events}
			runner := Runner{
				ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
				SandboxFactory: func(string, string) (sandboxintegration.Sandbox, error) { return sandbox, nil },
				AgentFactory:   factory,
				Executor:       &recordingExecutor{events: &events},
			}

			_, err := runner.Run(context.Background(), testDefinition())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "create model agent")
		})
	}
}

func TestRunnerRequiresAgentFactoryForBaselineRun(t *testing.T) {
	events := []string{}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) {
			events = append(events, "factory")
			return &fakeCluster{}, nil
		},
		Executor: &recordingExecutor{events: &events},
	}

	_, err := runner.Run(context.Background(), testDefinition())
	assert.EqualError(t, err, "baseline run requires a model agent factory")
	assert.Empty(t, events)
}

func TestRunnerRejectsAgentOnValidationRun(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{events: &events, kubeconfigPath: "/tmp/test.kubeconfig"}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		AgentFactory: rootagent.Factory(func(sandboxintegration.Executor) (rootagent.Agent, error) {
			return &fakeAgent{}, nil
		}),
		Executor: &recordingExecutor{events: &events},
	}

	_, err := runner.RunWithRepair(context.Background(), testDefinition(), scenario.Step{{Program: "repair"}})
	assert.ErrorContains(t, err, "validation run cannot use a model agent")
	assert.Empty(t, events)
}

func TestRunnerSupportsScenarioWithoutFaultInjection(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:         &events,
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}
	definition := testDefinition()
	definition.InjectFault = nil
	definition.VerifyFault = nil

	_, err := runner.RunWithRepair(context.Background(), definition, scenario.Step{{Program: "implement"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"create", "kubectl", "verify-clean", "implement", "verify-restored", "delete"}, events)
}

func TestRunnerSupportsVerifyOnlyFaultPhase(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:         &events,
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}
	definition := testDefinition()
	definition.InjectFault = nil

	_, err := runner.RunWithRepair(context.Background(), definition, scenario.Step{{Program: "repair"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"create", "kubectl", "verify-clean", "verify-fault", "repair", "verify-restored", "delete"}, events)
}

func TestRunnerRunsValidationRepairBeforeGrading(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:         &events,
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}

	_, err := runner.RunWithRepair(context.Background(), testDefinition(), scenario.Step{{Program: "repair"}})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"create", "kubectl", "verify-clean", "inject-fault", "verify-fault", "repair", "verify-restored", "delete",
	}, events)
}

func TestRunnerPreservesFailedStepEvidence(t *testing.T) {
	events := []string{}
	cluster := &fakeCluster{
		events:         &events,
		kubeconfigPath: "/tmp/test.kubeconfig",
		kubeconfigCtx:  "kind-test",
	}
	wantErr := errors.New("verification failed")
	executor := &recordingExecutor{
		events:   &events,
		failAt:   2,
		failWith: wantErr,
		results:  []command.Result{{}, {Stdout: "diagnostic", Stderr: "unhealthy", ExitCode: 7, Duration: 2 * time.Millisecond}},
	}
	runner := Runner{
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		Executor:       executor,
	}

	result, err := runner.RunWithRepair(context.Background(), testDefinition(), nil)

	assert.ErrorIs(t, err, wantErr)
	require.NotNil(t, result.Failure)
	assert.Equal(t, "verify clean", result.Failure.Phase)
	assert.Equal(t, 1, result.Failure.CommandIndex)
	assert.Equal(t, "verify-clean", result.Failure.Program)
	assert.Equal(t, "diagnostic", result.Failure.Stdout)
	assert.Equal(t, "unhealthy", result.Failure.Stderr)
	assert.Equal(t, 7, result.Failure.ExitCode)
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

func TestClusterNameFitsKindNodeNameLimit(t *testing.T) {
	name := clusterNameFor(strings.Repeat("a", 32))
	assert.LessOrEqual(t, len(name)+len("-control-plane"), 63)
}

func TestClusterNamesHaveDifferentRandomSuffixes(t *testing.T) {
	first := clusterNameFor("scenario")
	second := clusterNameFor("scenario")
	assert.NotEqual(t, first, second)
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
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}

	_, err := runner.RunWithRepair(context.Background(), testDefinition(), nil)
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
		ClusterFactory: func(string) (clusterintegration.Cluster, error) { return cluster, nil },
		Executor:       &recordingExecutor{events: &events},
	}

	_, err := runner.RunWithRepair(context.Background(), testDefinition(), nil)
	assert.ErrorIs(t, err, cleanupErr)
}
