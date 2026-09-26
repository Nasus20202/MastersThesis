package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRejectsInvalidArgumentsBeforeExecution(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing mode", want: "scenario, validation or corpus path is required"},
		{name: "both modes", args: []string{"--scenario", "scenario.yaml", "--validate", "validation.yaml"}, want: "cannot be combined"},
		{name: "corpus with scenario", args: []string{"--scenario", "scenario.yaml", "--check-corpus", "scenarios"}, want: "cannot be combined"},
		{name: "invalid parallelism", args: []string{"--scenario", "scenario.yaml", "--parallel", "0"}, want: "parallel must be at least 1"},
		{name: "invalid repeat", args: []string{"--scenario", "scenario.yaml", "--repeat", "0"}, want: "repeat must be at least 1"},
		{name: "invalid agent", args: []string{"--scenario", "scenario.yaml", "--agent", "unknown"}, want: "unsupported agent"},
		{name: "agent with validation", args: []string{"--validate", "validation.yaml", "--agent", "prompt"}, want: "only supported with scenario runs"},
		{name: "unexpected argument", args: []string{"--scenario", "scenario.yaml", "unexpected"}, want: "unexpected arguments"},
		{name: "blank path", args: []string{"--scenario", ""}, want: "path must not be blank"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			err := run(context.Background(), test.args, &logs)
			require.Error(t, err)
			assert.ErrorContains(t, err, test.want)
		})
	}
}

func TestRunHelpListsModesAndParameters(t *testing.T) {
	var logs bytes.Buffer
	err := run(context.Background(), []string{"--help"}, &logs)

	assert.ErrorIs(t, err, flag.ErrHelp)
	assert.Contains(t, logs.String(), "benchmark --config PATH ... --scenario PATH")
	assert.Contains(t, logs.String(), "NAME[,NAME]")
	assert.Contains(t, logs.String(), "benchmark --config PATH ... --validate PATH")
	assert.Contains(t, logs.String(), "benchmark --check-corpus PATH")
	assert.Contains(t, logs.String(), "-config")
	assert.Contains(t, logs.String(), "-agent")
	assert.Contains(t, logs.String(), "-repeat")
}

func TestStringList(t *testing.T) {
	var paths stringList
	require.NoError(t, paths.Set("first"))
	require.NoError(t, paths.Set("second"))

	assert.Equal(t, "first,second", paths.String())
	assert.ErrorContains(t, paths.Set(" "), "path must not be blank")
}

func TestAgentList(t *testing.T) {
	var agents agentList
	require.NoError(t, agents.Set("baseline"))
	require.NoError(t, agents.Set("prompt"))
	require.NoError(t, agents.Set("skill"))

	assert.Equal(t, "baseline,prompt,skill", agents.String())
	assert.ErrorContains(t, agents.Set(" "), "agent must not be blank")
}

func TestExplicitAllAgentSelection(t *testing.T) {
	assert.True(t, explicitAllAgentSelection(agentList{" all "}))
	assert.False(t, explicitAllAgentSelection(agentList{"baseline,prompt"}))
	assert.False(t, explicitAllAgentSelection(agentList{"all", "all"}))
}

func TestNewLoggerRejectsInvalidConfig(t *testing.T) {
	_, err := newLogger(&bytes.Buffer{}, benchmarkconfig.LoggingConfig{Level: "trace"})
	assert.ErrorContains(t, err, "parse log level")

	_, err = newLogger(&bytes.Buffer{}, benchmarkconfig.LoggingConfig{Format: "xml"})
	assert.ErrorContains(t, err, "unsupported log format")
}

func TestNewLoggerReadsConfiguredColor(t *testing.T) {
	var output bytes.Buffer
	logger, err := newLogger(&output, benchmarkconfig.LoggingConfig{Color: "always"})
	require.NoError(t, err)

	logger.Info("visible")

	assert.Contains(t, output.String(), "\x1b[")
}

func TestScenarioIDs(t *testing.T) {
	definitions := []scenario.Definition{{ID: "first"}, {ID: "second"}}
	assert.Equal(t, []string{"first", "second"}, scenarioIDs(definitions))
}

func TestValidationScenarioIDsAreUniqueAndOrdered(t *testing.T) {
	cases := []validation.ValidationCase{
		{Scenario: scenario.Definition{ID: "first"}},
		{Scenario: scenario.Definition{ID: "first"}},
		{Scenario: scenario.Definition{ID: "second"}},
	}

	assert.Equal(t, []string{"first", "second"}, validationScenarioIDs(cases))
}

const corpusScenarioYAML = `
id: %s
title: Corpus scenario
task: Restore the workload.
prepare:
  - program: prepare
verify_clean:
  - program: verify-clean
grading:
  - id: ready
    weight: 1
    check:
      program: check
`

const corpusValidationYAML = `
scenarios:
  - scenario_file: scenario.yaml
    cases:
      - id: repaired
        expected_score: 1
        expected_full_success: true
`

func TestRunCorpusCheck(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, root string)
		want    string
	}{
		{
			name: "valid pair",
			prepare: func(t *testing.T, root string) {
				t.Helper()
				writeCorpusScenario(t, filepath.Join(root, "good"), "good")
				writeCorpusValidation(t, filepath.Join(root, "good"))
			},
		},
		{
			name: "missing validation",
			prepare: func(t *testing.T, root string) {
				t.Helper()
				writeCorpusScenario(t, filepath.Join(root, "good"), "good")
			},
			want: "missing validation.yaml",
		},
		{
			name: "missing scenario",
			prepare: func(t *testing.T, root string) {
				t.Helper()
				writeCorpusScenario(t, filepath.Join(root, "good"), "good")
				writeCorpusValidation(t, filepath.Join(root, "good"))
				writeCorpusValidation(t, filepath.Join(root, "orphan"))
			},
			want: "missing scenario.yaml",
		},
		{
			name: "id mismatch",
			prepare: func(t *testing.T, root string) {
				t.Helper()
				writeCorpusScenario(t, filepath.Join(root, "wrong"), "good")
				writeCorpusValidation(t, filepath.Join(root, "wrong"))
			},
			want: "does not match directory name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			test.prepare(t, root)

			err := runCorpusCheck(root)
			if test.want == "" {
				require.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, test.want)
		})
	}
}

func writeCorpusScenario(t *testing.T, dir, id string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte(fmt.Sprintf(corpusScenarioYAML, id)), 0o600))
}

func writeCorpusValidation(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "validation.yaml"), []byte(corpusValidationYAML), 0o600))
}
