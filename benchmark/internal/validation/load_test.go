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

const validScenarioYAML = `
id: test-scenario
title: Test scenario
task: Restore the workload.
prepare:
  - program: prepare
verify_clean:
  - program: verify-clean
inject_fault:
  - program: inject-fault
verify_fault:
  - program: verify-fault
grading:
  - id: ready
    weight: 1
    check:
      program: check
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

func TestLoadInputsDiscoversValidationFilesInDirectories(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	require.NoError(t, os.MkdirAll(nested, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "validation.yml"), []byte(validValidationYAML), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "scenario.yaml"), []byte("id: unrelated"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "notes.yml"), []byte("notes: unrelated"), 0o600))

	definitions, err := LoadInputs([]string{root})
	require.NoError(t, err)
	require.Len(t, definitions, 1)
	assert.Equal(t, filepath.Join(nested, "../image-pull-failure/scenario.yaml"), definitions[0].Scenarios[0].ScenarioPath())
}

func TestLoadInputsReportsInvalidDiscoveredFile(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "validation.yaml"), []byte("scenarios: [invalid"), 0o600))

	_, err := LoadInputs([]string{root})
	assert.Error(t, err)
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

func TestLoadCasesLoadsScenariosAndPropagatesRepairDirectory(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "scenario.yaml"), []byte(validScenarioYAML), 0o600))
	validationPath := filepath.Join(root, "validation.yaml")
	validationYAML := `
scenarios:
  - scenario_file: scenario.yaml
    cases:
      - id: repaired
        repair:
          - program: repair
        expected_score: 1
        expected_full_success: true
`
	require.NoError(t, os.WriteFile(validationPath, []byte(validationYAML), 0o600))

	cases, err := LoadCases([]string{validationPath})
	require.NoError(t, err)
	require.Len(t, cases, 1)
	assert.Equal(t, "test-scenario", cases[0].Scenario.ID)
	assert.Equal(t, "repaired", cases[0].ID)
	assert.Equal(t, float64(1), cases[0].ExpectedScore)
	assert.True(t, cases[0].ExpectedFullSuccess)
	require.Len(t, cases[0].Repair, 1)
	assert.Equal(t, filepath.Dir(validationPath), cases[0].Repair[0].Spec().Dir)
}

func TestLoadCasesRejectsDuplicateCasesAcrossInputs(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "scenario.yaml"), []byte(validScenarioYAML), 0o600))
	validationYAML := `
scenarios:
  - scenario_file: scenario.yaml
    cases:
      - id: repaired
        expected_score: 1
        expected_full_success: true
`
	first := filepath.Join(root, "first.yaml")
	second := filepath.Join(root, "second.yaml")
	require.NoError(t, os.WriteFile(first, []byte(validationYAML), 0o600))
	require.NoError(t, os.WriteFile(second, []byte(validationYAML), 0o600))

	_, err := LoadCases([]string{first, second})
	assert.ErrorContains(t, err, `duplicate validation case "test-scenario/repaired"`)
}
