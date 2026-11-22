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
	"slices"
	"strings"
	"syscall"
	"time"

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/logging"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
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
	flags.Var(&scenarioPaths, "scenario", "path to a scenario YAML file or directory; may be repeated or comma-separated (default: the records' scenarios with --continue-from)")
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
	var continueFrom stringList
	flags.Var(&continueFrom, "continue-from", "recorded attempt to continue in a fresh cluster, one per scenario; may be repeated or comma-separated, and names the scenario when --scenario is omitted (needs one agent)")
	list := flags.Bool("list", false, "print the scenario files selected by --scenario and --tag, one per line, and exit")
	var mergeIDs stringList
	flags.Var(&mergeIDs, "merge", "merge runs of one configuration, such as benchmark shards or repeated runs, into a new run and print its ID; may be repeated")
	checkCorpus := flags.String("check-corpus", "", "validate the scenario corpus under PATH and exit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	scenarioPaths = splitList(scenarioPaths)
	continueFrom = splitList(continueFrom)
	var prefixes map[string]*common.Prefix
	if len(continueFrom) > 0 {
		var err error
		prefixes, err = loadPrefixes(continueFrom)
		if err != nil {
			return err
		}
	}
	if len(scenarioPaths) == 0 && len(prefixes) > 0 {
		root, err := scenariosRoot()
		if err != nil {
			return err
		}
		scenarioPaths, err = scenarioPathsFor(root, prefixes)
		if err != nil {
			return err
		}
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
	if *list {
		if len(scenarioPaths) == 0 {
			return errors.New("list requires scenario paths")
		}
		var allowed map[string]bool
		if len(prefixes) > 0 {
			allowed = make(map[string]bool, len(prefixes))
			for scenario := range prefixes {
				allowed[scenario] = true
			}
		}
		return listScenarios(output, scenarioPaths, tagFilter, allowed)
	}

	options := runOptions{parallel: *parallel, repeat: *repeat, tags: tagFilter, config: benchmarkConfig, terminal: terminal}
	if len(prefixes) > 0 {
		if len(validationPaths) > 0 || strings.TrimSpace(*resumeID) != "" || len(agentNames) != 1 {
			return errors.New("continue-from needs a scenario run with exactly one agent, and no resume")
		}
		options.continueFroms = prefixes
	}
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

// scenarioPathsFor is the scenario file of each continuation record, for a run
// that names no scenario paths itself.
func scenarioPathsFor(root string, prefixes map[string]*common.Prefix) (stringList, error) {
	definitions, err := scenario.LoadInputs([]string{root})
	if err != nil {
		return nil, err
	}
	paths := make(stringList, 0, len(prefixes))
	seen := make(map[string]bool, len(prefixes))
	for _, definition := range definitions {
		if _, ok := prefixes[definition.ID]; ok {
			paths = append(paths, definition.Path)
			seen[definition.ID] = true
		}
	}
	for id := range prefixes {
		if !seen[id] {
			return nil, fmt.Errorf("no scenario %q in %s", id, root)
		}
	}
	return paths, nil
}

// listScenarios prints the scenario files a run with the same paths and tag
// filter would include, so callers can split them into shards.
func listScenarios(output io.Writer, inputs []string, tags scenario.TagFilter, allowed map[string]bool) error {
	definitions, err := scenario.LoadInputs(inputs)
	if err != nil {
		return err
	}
	definitions = scenario.FilterByTags(definitions, tags)
	if allowed != nil {
		definitions = slices.DeleteFunc(slices.Clone(definitions), func(d scenario.Definition) bool {
			return !allowed[d.ID]
		})
	}
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

// splitList expands the values of a flag that accepts a repeated or
// comma-separated list, the way --agent and --tag do.
func splitList(values stringList) stringList {
	expanded := make(stringList, 0, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			if item = strings.TrimSpace(item); item != "" {
				expanded = append(expanded, item)
			}
		}
	}
	return expanded
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
