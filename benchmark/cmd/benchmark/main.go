// Package main is the benchmark CLI: it parses flags and configuration, then
// runs or validates scenario-based agent benchmark attempts.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/logging"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

const envNoColor = "NO_COLOR"

func main() {
	os.Exit(runMain())
}

func runMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		slog.Error("benchmark failed", "error", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, args []string, output, logOutput io.Writer) error {
	terminal := ui.NewTerminal(logOutput)
	flags := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	flags.SetOutput(logOutput)
	flags.Usage = func() {
		fmt.Fprintln(logOutput, "Usage:")
		fmt.Fprintln(logOutput, "  benchmark --config PATH ... --scenario PATH [--tag KEY=VALUE] [--agent NAME[,NAME] ...] [--parallel N] [--repeat N] [--resume RUN_ID]")
		fmt.Fprintln(logOutput, "  benchmark --config PATH ... --validate PATH [--tag KEY=VALUE] [--parallel N] [--repeat N]")
		fmt.Fprintln(logOutput, "  benchmark --scenario PATH [--tag KEY=VALUE] --list")
		fmt.Fprintln(logOutput, "  benchmark --merge RUN_ID ...")
		fmt.Fprintln(logOutput, "  benchmark --check-corpus PATH")
		fmt.Fprintln(logOutput, "\nOptions:")
		flags.PrintDefaults()
	}
	var scenarioPaths stringList
	flags.Var(&scenarioPaths, "scenario", "path to a scenario YAML file or directory; may be repeated")
	var validationPaths stringList
	flags.Var(&validationPaths, "validate", "path to a validation YAML file or directory; may be repeated")
	var configPaths stringList
	flags.Var(&configPaths, "config", "load benchmark YAML configuration; may be repeated in overlay order")
	var agentValues stringList
	flags.Var(&agentValues, "agent", "benchmark agent(s): all, baseline, prompt, skill, or rag; may be repeated or comma-separated (default: all)")
	var tagValues stringList
	flags.Var(&tagValues, "tag", "filter scenarios by tag, e.g. difficulty=hard or area=networking; may be repeated or comma-separated")
	parallel := flags.Int("parallel", 1, "maximum number of tasks running at once")
	repeat := flags.Int("repeat", 1, "number of times to run each scenario or validation case")
	resumeID := flags.String("resume", "", "resume an incomplete scenario run by ID")
	list := flags.Bool("list", false, "print the scenario files selected by --scenario and --tag, one per line, and exit")
	var mergeIDs stringList
	flags.Var(&mergeIDs, "merge", "merge runs of one configuration, such as benchmark shards or repeated runs, into a new run and print its ID; may be repeated")
	checkCorpus := flags.String("check-corpus", "", "validate the scenario corpus under PATH and exit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	corpusPath := strings.TrimSpace(*checkCorpus)
	modes := 0
	for _, selected := range []bool{len(scenarioPaths) > 0, len(validationPaths) > 0, len(mergeIDs) > 0, corpusPath != ""} {
		if selected {
			modes++
		}
	}
	if modes > 1 {
		return errors.New("scenario, validate, merge and check-corpus cannot be combined")
	}
	if modes == 0 {
		return errors.New("scenario, validation, merge or corpus path is required; use --scenario, --validate, --merge or --check-corpus")
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	tagFilter, err := scenario.ParseTagFilter(tagValues)
	if err != nil {
		return err
	}
	benchmarkConfig, err := config.Load(configPaths...)
	if err != nil {
		return err
	}
	logger, err := newLogger(terminal, benchmarkConfig.Logging)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)
	if *list {
		if len(scenarioPaths) == 0 {
			return errors.New("list requires scenario paths")
		}
		return listScenarios(output, scenarioPaths, tagFilter)
	}
	if len(mergeIDs) > 0 {
		return mergeRuns(output, mergeIDs)
	}
	if corpusPath != "" {
		if !tagFilter.Empty() {
			return errors.New("tag filtering is only supported with scenario or validation runs")
		}
		return runCorpusCheck(corpusPath)
	}
	if *parallel < 1 {
		return errors.New("parallel must be at least 1")
	}
	if *repeat < 1 {
		return errors.New("repeat must be at least 1")
	}
	agentNames, err := commandagent.Select(agentValues...)
	if err != nil {
		return err
	}

	options := runOptions{parallel: *parallel, repeat: *repeat, tags: tagFilter, config: benchmarkConfig, terminal: terminal}
	if len(validationPaths) > 0 {
		if strings.TrimSpace(*resumeID) != "" {
			return errors.New("resume is only supported with scenario runs")
		}
		if len(agentValues) > 0 && !explicitAllAgentSelection(agentValues) {
			return errors.New("agent selection is only supported with scenario runs")
		}
		return runValidation(ctx, options, validationPaths)
	}
	return runBenchmark(ctx, options, scenarioPaths, agentNames, strings.TrimSpace(*resumeID))
}

// listScenarios prints the scenario files a run with the same paths and tag
// filter would include, so callers can split them into shards.
func listScenarios(output io.Writer, inputs []string, tags scenario.TagFilter) error {
	definitions, err := scenario.LoadInputs(inputs)
	if err != nil {
		return err
	}
	definitions = scenario.FilterByTags(definitions, tags)
	if len(definitions) == 0 {
		return fmt.Errorf("no scenarios match tag filter %q", tags.String())
	}
	for _, definition := range definitions {
		if _, err := fmt.Fprintln(output, definition.Path); err != nil {
			return err
		}
	}
	return nil
}

func mergeRuns(output io.Writer, runIDs []string) error {
	metadata, err := results.Merge(resultsDir, results.NewRunID(time.Now()), runIDs)
	if err != nil {
		return err
	}
	slog.Info("runs merged", "run_id", metadata.RunID, "runs", len(runIDs), "state", metadata.State)
	_, err = fmt.Fprintln(output, metadata.RunID)
	return err
}

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("value must not be blank")
	}
	*s = append(*s, value)
	return nil
}

func explicitAllAgentSelection(values stringList) bool {
	return len(values) == 1 && strings.EqualFold(strings.TrimSpace(values[0]), string(commandagent.All))
}

func newLogger(output io.Writer, settings config.LoggingConfig) (*slog.Logger, error) {
	level, err := logging.ParseLevel(settings.Level)
	if err != nil {
		return nil, err
	}
	return logging.New(output, logging.Format(settings.Format), level, logColor(output, settings.Color))
}

// logColor applies the configured color setting, then NO_COLOR, then whether
// the output is an interactive terminal.
func logColor(output io.Writer, setting string) bool {
	switch strings.ToLower(strings.TrimSpace(setting)) {
	case "always", "true", "1":
		return true
	case "never", "false", "0":
		return false
	}
	if os.Getenv(envNoColor) != "" {
		return false
	}
	terminal, ok := output.(interface{ ColorEnabled() bool })
	return ok && terminal.ColorEnabled()
}
