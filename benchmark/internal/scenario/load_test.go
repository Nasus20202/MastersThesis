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
