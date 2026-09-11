package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validValidationYAML = `
scenarios:
  - scenario_file: ../image-pull-failure/scenario.yaml
    cases:
      - id: broken
        expected_score: 0
        expected_full_success: false
      - id: repaired
        repair:
          - program: kubectl
            args: [set, image, deployment/app, app=nginx:1.31.5]
        expected_score: 1
        expected_full_success: true
`

func TestLoadReadsValidationManifest(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "validation.yaml")
	require.NoError(t, os.WriteFile(path, []byte(validValidationYAML), 0o600))

	definition, err := Load(path)
	require.NoError(t, err)
	require.Len(t, definition.Scenarios, 1)
	assert.Equal(t, filepath.Join(root, "../image-pull-failure/scenario.yaml"), definition.Scenarios[0].ScenarioPath())
	assert.Equal(t, filepath.Dir(path), definition.Scenarios[0].Cases[1].Repair[0].Spec().Dir)
	require.NotNil(t, definition.Scenarios[0].Cases[0].ExpectedScore)
	assert.Equal(t, float64(0), *definition.Scenarios[0].Cases[0].ExpectedScore)
}

func TestLoadInputsRecursivelyDiscoversYAMLFiles(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	require.NoError(t, os.MkdirAll(nested, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "checks.yml"), []byte(validValidationYAML), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "scenario.yaml"), []byte("id: unrelated"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "notes.yml"), []byte("notes: unrelated"), 0o600))

	definitions, err := LoadInputs([]string{root})
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	assert.Equal(t, filepath.Join(nested, "../image-pull-failure/scenario.yaml"), definitions[0].Scenarios[0].ScenarioPath())
}

func TestLoadInputsExplicitInvalidFileIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "validation.yaml")
	require.NoError(t, os.WriteFile(path, []byte("scenarios: [invalid"), 0o600))

	_, err := LoadInputs([]string{path})
	assert.Error(t, err)
}

func TestParseRejectsDuplicateCaseIDs(t *testing.T) {
	data := strings.Replace(validValidationYAML, "      - id: repaired", "      - id: broken", 1)

	_, err := Parse([]byte(data))
	assert.ErrorContains(t, err, "duplicate case id")
}

func TestParseRequiresExpectedValues(t *testing.T) {
	data := strings.Replace(validValidationYAML, "        expected_score: 0\n", "", 1)

	_, err := Parse([]byte(data))
	assert.ErrorContains(t, err, "ExpectedScore")
}
