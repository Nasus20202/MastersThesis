package scenario

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsScenarioFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(validScenarioYAML), 0o600); err != nil {
		t.Fatalf("write scenario: %v", err)
	}

	definition, err := Load(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	if definition.Title != "Test scenario" {
		t.Fatalf("title = %q, want Test scenario", definition.Title)
	}
	if got, want := definition.Prepare[0].Spec().Dir, filepath.Dir(path); got != want {
		t.Fatalf("command directory = %q, want %q", got, want)
	}
}

func TestLoadResolvesKindConfigPath(t *testing.T) {
	data := strings.Replace(validScenarioYAML, "task: Restore the test workload.", "task: Restore the test workload.\n\ncluster:\n  kind:\n    config: kind/cluster.yaml", 1)
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("write scenario: %v", err)
	}

	definition, err := Load(path)
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	want := filepath.Join(filepath.Dir(path), "kind", "cluster.yaml")
	if got := definition.Cluster.Kind.ConfigPath(); got != want {
		t.Fatalf("kind config path = %q, want %q", got, want)
	}
}
