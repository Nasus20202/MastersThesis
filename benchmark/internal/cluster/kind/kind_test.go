package kind

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
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
	if _, err := New(nil, Config{Name: "cluster"}); err == nil {
		t.Fatal("expected missing executor error")
	}
	if _, err := New(&fakeExecutor{}, Config{Name: " "}); err == nil {
		t.Fatal("expected missing name error")
	}
	if _, err := New(&fakeExecutor{}, Config{Name: "cluster", ConfigPath: " "}); err == nil {
		t.Fatal("expected invalid config path error")
	}
}

func TestClusterRunsCreateAndDelete(t *testing.T) {
	executor := &fakeExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := cluster.Create(context.Background()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := cluster.Delete(context.Background()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if len(executor.specs) != 2 {
		t.Fatalf("got %d command specs, want 2", len(executor.specs))
	}
	if got := executor.specs[0]; got.Program != program || !slices.Equal(got.Args, []string{"create", "cluster", "--name", "benchmark"}) {
		t.Fatalf("create spec = %#v", got)
	}
	if got := executor.specs[1]; got.Program != program || !slices.Equal(got.Args, []string{"delete", "cluster", "--name", "benchmark"}) {
		t.Fatalf("delete spec = %#v", got)
	}
	if got := cluster.Context(); got != "kind-benchmark" {
		t.Fatalf("Context() = %q, want %q", got, "kind-benchmark")
	}
}

func TestClusterUsesConfigFile(t *testing.T) {
	executor := &fakeExecutor{}
	cluster, err := New(executor, Config{Name: "benchmark", ConfigPath: "/tmp/kind.yaml"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := cluster.Create(context.Background()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if got := executor.specs[0]; got.Program != program || got.Dir != "/tmp" || !slices.Equal(got.Args, []string{"create", "cluster", "--name", "benchmark", "--config", "/tmp/kind.yaml"}) {
		t.Fatalf("create spec = %#v", got)
	}
}

func TestClusterReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("kind failed")
	cluster, err := New(&fakeExecutor{err: wantErr}, Config{Name: "benchmark"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = cluster.Create(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want wrapped %v", err, wantErr)
	}
}
