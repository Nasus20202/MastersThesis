package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
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
		{name: "missing mode", want: "scenario, validation, merge or corpus path is required"},
		{name: "both modes", args: []string{"--scenario", "scenario.yaml", "--validate", "validation.yaml"}, want: "cannot be combined"},
		{name: "corpus with scenario", args: []string{"--scenario", "scenario.yaml", "--check-corpus", "scenarios"}, want: "cannot be combined"},
		{name: "merge with scenario", args: []string{"--scenario", "scenario.yaml", "--merge", "run-1"}, want: "cannot be combined"},
		{name: "invalid parallelism", args: []string{"--scenario", "scenario.yaml", "--parallel", "0"}, want: "parallel must be at least 1"},
		{name: "invalid repeat", args: []string{"--scenario", "scenario.yaml", "--repeat", "0"}, want: "repeat must be at least 1"},
		{name: "invalid agent", args: []string{"--scenario", "scenario.yaml", "--agent", "unknown"}, want: "unsupported agent"},
		{name: "agent with validation", args: []string{"--validate", "validation.yaml", "--agent", "prompt"}, want: "only supported with scenario runs"},
		{name: "unexpected argument", args: []string{"--scenario", "scenario.yaml", "unexpected"}, want: "unexpected arguments"},
		{name: "blank path", args: []string{"--scenario", ""}, want: "value must not be blank"},
		{name: "list without scenarios", args: []string{"--validate", "validation.yaml", "--list"}, want: "list requires scenario paths"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			err := run(context.Background(), test.args, io.Discard, &logs)
			require.Error(t, err)
			assert.ErrorContains(t, err, test.want)
		})
	}
}

func TestRunHelpListsModesAndParameters(t *testing.T) {
	var logs bytes.Buffer
	err := run(context.Background(), []string{"--help"}, io.Discard, &logs)

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
	var values stringList
	require.NoError(t, values.Set("first"))
	require.NoError(t, values.Set("second"))

	assert.Equal(t, "first,second", values.String())
	assert.ErrorContains(t, values.Set(" "), "value must not be blank")
}

func TestExplicitAllAgentSelection(t *testing.T) {
	assert.True(t, explicitAllAgentSelection(stringList{" all "}))
	assert.False(t, explicitAllAgentSelection(stringList{"baseline,prompt"}))
	assert.False(t, explicitAllAgentSelection(stringList{"all", "all"}))
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

const corpusSourcesYAML = `sources:
  - path: docs/page.md
`

const corpusScenarioYAML = `
id: %s
title: Corpus scenario
task: Restore the workload.
` + corpusSourcesYAML + `prepare:
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
		{
			name: "missing sources",
			prepare: func(t *testing.T, root string) {
				t.Helper()
				dir := filepath.Join(root, "good")
				writeCorpusScenario(t, dir, "good")
				writeCorpusValidation(t, dir)
				path := filepath.Join(dir, "scenario.yaml")
				data, err := os.ReadFile(path)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(data), corpusSourcesYAML, "", 1)), 0o600))
			},
			want: "scenario has no sources",
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

func TestRunListPrintsScenariosMatchingTags(t *testing.T) {
	var output, logs bytes.Buffer
	err := run(context.Background(), []string{"--scenario", "../../scenarios", "--tag", "difficulty=hard", "--list"}, &output, &logs)
	require.NoError(t, err)

	paths := strings.Fields(output.String())
	require.NotEmpty(t, paths)
	all, err := scenario.LoadInputs([]string{"../../scenarios"})
	require.NoError(t, err)
	assert.Less(t, len(paths), len(all))
	for _, path := range paths {
		definition, err := scenario.Load(path)
		require.NoError(t, err)
		assert.Equal(t, "hard", definition.Tags["difficulty"], path)
	}

	err = run(context.Background(), []string{"--scenario", "../../scenarios", "--tag", "difficulty=none", "--list"}, &output, &logs)
	assert.ErrorContains(t, err, "no scenarios match")
}
