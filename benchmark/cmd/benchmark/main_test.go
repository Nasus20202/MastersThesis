package main

import (
	"bytes"
	"context"
	"flag"
	"testing"

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
		{name: "missing mode", want: "scenario or validation path is required"},
		{name: "both modes", args: []string{"--scenario", "scenario.yaml", "--validate", "validation.yaml"}, want: "cannot be combined"},
		{name: "invalid parallelism", args: []string{"--scenario", "scenario.yaml", "--parallel", "0"}, want: "parallel must be at least 1"},
		{name: "invalid repeat", args: []string{"--scenario", "scenario.yaml", "--repeat", "0"}, want: "repeat must be at least 1"},
		{name: "unknown agent", args: []string{"--scenario", "scenario.yaml", "--agent", "unknown"}, want: "unsupported agent"},
		{name: "agent with validation", args: []string{"--validate", "validation.yaml", "--agent", "baseline"}, want: "agent cannot be used with validation"},
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
	assert.Contains(t, logs.String(), "benchmark --scenario PATH")
	assert.Contains(t, logs.String(), "benchmark --validate PATH")
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
	require.NoError(t, agents.Set("skill"))
	assert.Equal(t, "baseline,skill", agents.String())
	assert.ErrorContains(t, agents.Set(" "), "agent name must not be blank")
}

func TestNewLoggerRejectsInvalidEnvironment(t *testing.T) {
	t.Setenv("BENCHMARK_LOG_LEVEL", "trace")
	t.Setenv("BENCHMARK_LOG_FORMAT", "")
	_, err := newLogger(&bytes.Buffer{})
	assert.ErrorContains(t, err, "parse log level")

	t.Setenv("BENCHMARK_LOG_LEVEL", "")
	t.Setenv("BENCHMARK_LOG_FORMAT", "xml")
	_, err = newLogger(&bytes.Buffer{})
	assert.ErrorContains(t, err, "unsupported log format")
}

func TestNewLoggerReadsColorEnvironmentAtApplicationBoundary(t *testing.T) {
	t.Setenv("BENCHMARK_LOG_COLOR", "always")
	var output bytes.Buffer
	logger, err := newLogger(&output)
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
