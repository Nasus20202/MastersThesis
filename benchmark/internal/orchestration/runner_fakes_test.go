package orchestration

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
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

type fakeSandbox struct {
	events   *[]string
	buildErr error
	startErr error
	stopErr  error
}

type fakeImageBuilder struct {
	events   *[]string
	buildErr error
}

func (b *fakeImageBuilder) Build(context.Context) error {
	*b.events = append(*b.events, "sandbox-image-build")
	return b.buildErr
}

func (s *fakeSandbox) Build(context.Context) error {
	*s.events = append(*s.events, "sandbox-build")
	return s.buildErr
}

func (s *fakeSandbox) Start(context.Context) error {
	*s.events = append(*s.events, "sandbox-start")
	return s.startErr
}

func (s *fakeSandbox) Stop(context.Context) error {
	*s.events = append(*s.events, "sandbox-stop")
	return s.stopErr
}

func (s *fakeSandbox) Exec(context.Context, command.Spec) (command.Result, error) {
	*s.events = append(*s.events, "sandbox-exec")
	return command.Result{}, nil
}

type fakeAgent struct {
	events *[]string
	task   string
	err    error
}

func (a *fakeAgent) Run(_ context.Context, task string) (common.Result, error) {
	*a.events = append(*a.events, "agent")
	a.task = task
	return common.Result{Task: task, Termination: "completed"}, a.err
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
		resultIndex := len(e.specs) - 1
		if resultIndex < len(e.results) {
			return e.results[resultIndex], e.failWith
		}
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
