package scenario

import (
	"strings"
	"testing"
)

const validScenarioYAML = `
id: test-scenario
title: Test scenario
task: Restore the test workload.

prepare:
  - program: prepare
    args: [manifest.yaml]

verify_clean:
  - program: verify-clean

inject_fault:
  - program: inject-fault

verify_fault:
  - program: verify-fault

reset:
  - program: reset
`

func TestParseValidScenario(t *testing.T) {
	definition, err := Parse([]byte(validScenarioYAML))
	if err != nil {
		t.Fatalf("parse scenario: %v", err)
	}

	if definition.ID != "test-scenario" {
		t.Fatalf("ID = %q, want test-scenario", definition.ID)
	}
	if len(definition.Prepare) != 1 || definition.Prepare[0].Program != "prepare" {
		t.Fatalf("prepare = %#v, want one prepare command", definition.Prepare)
	}
	if got := definition.Prepare[0].Spec().Args; len(got) != 1 || got[0] != "manifest.yaml" {
		t.Fatalf("prepare command args = %#v, want manifest.yaml", got)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	_, err := Parse([]byte(validScenarioYAML + "\nunknown: value\n"))
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("parse error = %v, want unknown-field error", err)
	}
}

func TestParseReportsValidationErrors(t *testing.T) {
	_, err := Parse([]byte("id: scenario\ntitle: Test\ntask: \" \"\n"))
	if err == nil || !strings.Contains(err.Error(), "Task") {
		t.Fatalf("validation error = %v, want Task validation error", err)
	}
}

func validDefinition(t *testing.T) Definition {
	t.Helper()

	definition, err := Parse([]byte(validScenarioYAML))
	if err != nil {
		t.Fatalf("parse scenario: %v", err)
	}
	return definition
}

func TestValidateRejectsMissingPhase(t *testing.T) {
	definition := validDefinition(t)
	definition.Prepare = nil

	if err := definition.Validate(); err == nil || !strings.Contains(err.Error(), "Prepare") {
		t.Fatalf("validation error = %v, want Prepare validation error", err)
	}
}

func TestValidateRejectsEmptyPhase(t *testing.T) {
	definition := validDefinition(t)
	definition.Prepare = Step{}

	if err := definition.Validate(); err == nil || !strings.Contains(err.Error(), "Prepare") {
		t.Fatalf("validation error = %v, want Prepare validation error", err)
	}
}

func TestValidateRejectsBlankProgram(t *testing.T) {
	definition := validDefinition(t)
	definition.Prepare = Step{{Program: "  "}}

	if err := definition.Validate(); err == nil || !strings.Contains(err.Error(), "Program") {
		t.Fatalf("validation error = %v, want Program validation error", err)
	}
}

func TestValidateRejectsInvalidID(t *testing.T) {
	definition := validDefinition(t)
	definition.ID = "Image Pull Failure"

	if err := definition.Validate(); err == nil || !strings.Contains(err.Error(), "scenarioid") {
		t.Fatalf("validation error = %v, want scenarioid validation error", err)
	}
}

func TestValidateRejectsLongID(t *testing.T) {
	definition := validDefinition(t)
	definition.ID = strings.Repeat("a", 33)

	if err := definition.Validate(); err == nil || !strings.Contains(err.Error(), "max") {
		t.Fatalf("validation error = %v, want max validation error", err)
	}
}

func TestParseRejectsUnknownNestedField(t *testing.T) {
	data := strings.Replace(validScenarioYAML, "    args: [manifest.yaml]", "    args: [manifest.yaml]\n    extra: value", 1)
	_, err := Parse([]byte(data))
	if err == nil || !strings.Contains(err.Error(), "extra") {
		t.Fatalf("parse error = %v, want nested unknown-field error", err)
	}
}

func TestParseRejectsMultipleDocuments(t *testing.T) {
	_, err := Parse([]byte(validScenarioYAML + "\n---\n" + validScenarioYAML))
	if err == nil || !strings.Contains(err.Error(), "exactly one document") {
		t.Fatalf("parse error = %v, want multiple-document error", err)
	}
}
