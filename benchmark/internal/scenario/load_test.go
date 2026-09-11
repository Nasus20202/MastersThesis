package scenario

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadReadsScenarioFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	require.NoError(t, os.WriteFile(path, []byte(validScenarioYAML), 0o600))

	definition, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "Test scenario", definition.Title)
	assert.Equal(t, filepath.Dir(path), definition.Prepare[0].Spec().Dir)
	assert.Equal(t, filepath.Dir(path), definition.Grading[0].Check.Spec().Dir)
}

func TestLoadResolvesKindConfigPath(t *testing.T) {
	data := strings.Replace(validScenarioYAML, "task: Restore the test workload.", "task: Restore the test workload.\n\ncluster:\n  kind:\n    config: kind/cluster.yaml", 1)
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	require.NoError(t, os.WriteFile(path, []byte(data), 0o600))

	definition, err := Load(path)
	require.NoError(t, err)

	want := filepath.Join(filepath.Dir(path), "kind", "cluster.yaml")
	assert.Equal(t, want, definition.Cluster.Kind.ConfigPath())
}

func TestLoadInputsRecursivelyDiscoversSortedScenarios(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested"), 0o700))
	writeScenario(t, filepath.Join(root, "nested", "b.yml"), "second-scenario")
	writeScenario(t, filepath.Join(root, "a.yaml"), "first-scenario")
	require.NoError(t, os.WriteFile(filepath.Join(root, "notes.yaml"), []byte("title: unrelated"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ignored.txt"), []byte(validScenarioYAML), 0o600))

	definitions, err := LoadInputs([]string{root})
	require.NoError(t, err)
	require.Len(t, definitions, 2)
	assert.Equal(t, "first-scenario", definitions[0].ID)
	assert.Equal(t, "second-scenario", definitions[1].ID)
}

func TestLoadInputsExplicitInvalidFileIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.yaml")
	require.NoError(t, os.WriteFile(path, []byte("id: invalid"), 0o600))

	_, err := LoadInputs([]string{path})
	assert.Error(t, err)
}

func TestLoadInputsSkipsInvalidDiscoveredFile(t *testing.T) {
	root := t.TempDir()
	writeScenario(t, filepath.Join(root, "valid.yaml"), "valid-scenario")
	require.NoError(t, os.WriteFile(filepath.Join(root, "invalid.yaml"), []byte("id: invalid"), 0o600))

	definitions, err := LoadInputs([]string{root})
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	assert.Equal(t, "valid-scenario", definitions[0].ID)
}

func TestLoadInputsRejectsDuplicateScenarioIDs(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.yaml")
	second := filepath.Join(root, "second.yaml")
	writeScenario(t, first, "same-scenario")
	writeScenario(t, second, "same-scenario")

	_, err := LoadInputs([]string{first, second})
	assert.ErrorContains(t, err, `duplicate scenario id "same-scenario"`)
}

func TestLoadInputsRequiresPath(t *testing.T) {
	_, err := LoadInputs(nil)
	assert.ErrorContains(t, err, "at least one scenario path")
}

func writeScenario(t *testing.T, path, id string) {
	t.Helper()
	data := strings.Replace(validScenarioYAML, "id: test-scenario", "id: "+id, 1)
	require.NoError(t, os.WriteFile(path, []byte(data), 0o600))
}
