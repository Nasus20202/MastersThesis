package scenario

import (
	"os"
	"path/filepath"
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
