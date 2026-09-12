package scenario

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validScenarioYAML = `
id: test-scenario
title: Test scenario
task: Restore the test workload.

prepare:
  - program: prepare
    args: [manifest.yaml]
    env:
      CHECK_MODE: strict

verify_clean:
  - program: verify-clean

inject_fault:
  - program: inject-fault

verify_fault:
  - program: verify-fault

grading:
  - id: workload-restored
    weight: 1
    check:
      program: verify-restored
`

func TestParseValidScenario(t *testing.T) {
	definition, err := Parse([]byte(validScenarioYAML))
	require.NoError(t, err)

	assert.Equal(t, "test-scenario", definition.ID)
	require.Len(t, definition.Prepare, 1)
	assert.Equal(t, "prepare", definition.Prepare[0].Program)
	assert.Equal(t, []string{"manifest.yaml"}, definition.Prepare[0].Spec().Args)
	assert.Equal(t, "strict", definition.Prepare[0].Spec().Env["CHECK_MODE"])
}

func TestParseRejectsUnknownField(t *testing.T) {
	_, err := Parse([]byte(validScenarioYAML + "\nunknown: value\n"))
	assert.ErrorContains(t, err, "unknown")
}

func TestParseReportsValidationErrors(t *testing.T) {
	_, err := Parse([]byte("id: scenario\ntitle: Test\ntask: \" \"\n"))
	assert.ErrorContains(t, err, "Task")
}

func validDefinition(t *testing.T) Definition {
	t.Helper()

	definition, err := Parse([]byte(validScenarioYAML))
	require.NoError(t, err)
	return definition
}

func TestValidateAllowsScenarioWithoutFault(t *testing.T) {
	definition := validDefinition(t)
	definition.InjectFault = nil
	definition.VerifyFault = nil

	assert.NoError(t, definition.Validate())
}

func TestValidateAllowsIndependentFaultPhases(t *testing.T) {
	t.Run("inject only", func(t *testing.T) {
		definition := validDefinition(t)
		definition.VerifyFault = nil

		assert.NoError(t, definition.Validate())
	})

	t.Run("verify only", func(t *testing.T) {
		definition := validDefinition(t)
		definition.InjectFault = nil

		assert.NoError(t, definition.Validate())
	})
}

func TestValidateRejectsMissingPhase(t *testing.T) {
	definition := validDefinition(t)
	definition.Prepare = nil

	assert.ErrorContains(t, definition.Validate(), "Prepare")
}

func TestValidateRejectsEmptyPhase(t *testing.T) {
	definition := validDefinition(t)
	definition.Prepare = Step{}

	assert.ErrorContains(t, definition.Validate(), "Prepare")
}

func TestValidateRejectsBlankProgram(t *testing.T) {
	definition := validDefinition(t)
	definition.Prepare = Step{{Program: "  "}}

	assert.ErrorContains(t, definition.Validate(), "Program")
}

func TestValidateRejectsInvalidID(t *testing.T) {
	definition := validDefinition(t)
	definition.ID = "Image Pull Failure"

	assert.ErrorContains(t, definition.Validate(), "scenarioid")
}

func TestValidateRejectsInvalidCriterion(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Definition)
		want   string
	}{
		{
			name: "blank id",
			change: func(definition *Definition) {
				definition.Grading[0].ID = " "
			},
			want: "ID",
		},
		{
			name: "zero weight",
			change: func(definition *Definition) {
				definition.Grading[0].Weight = 0
			},
			want: "Weight",
		},
		{
			name: "blank check program",
			change: func(definition *Definition) {
				definition.Grading[0].Check.Program = " "
			},
			want: "Program",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validDefinition(t)
			test.change(&definition)

			assert.ErrorContains(t, definition.Validate(), test.want)
		})
	}
}

func TestValidateRejectsMissingGrading(t *testing.T) {
	definition := validDefinition(t)
	definition.Grading = nil

	assert.ErrorContains(t, definition.Validate(), "Grading")
}

func TestValidateRejectsDuplicateCriterionIDs(t *testing.T) {
	definition := validDefinition(t)
	definition.Grading = append(definition.Grading, definition.Grading[0])

	assert.ErrorContains(t, definition.Validate(), "duplicate grading criterion id")
}

func TestValidateRejectsLongID(t *testing.T) {
	definition := validDefinition(t)
	definition.ID = strings.Repeat("a", 33)

	assert.ErrorContains(t, definition.Validate(), "max")
}

func TestParseRejectsUnknownNestedField(t *testing.T) {
	data := strings.Replace(validScenarioYAML, "    args: [manifest.yaml]", "    args: [manifest.yaml]\n    extra: value", 1)
	_, err := Parse([]byte(data))
	assert.ErrorContains(t, err, "extra")
}

func TestParseRejectsMultipleDocuments(t *testing.T) {
	_, err := Parse([]byte(validScenarioYAML + "\n---\n" + validScenarioYAML))
	assert.ErrorContains(t, err, "exactly one document")
}
