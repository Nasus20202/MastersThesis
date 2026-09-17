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

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/logging"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/ui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		slog.Error("benchmark failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, logOutput io.Writer) error {
	terminal := ui.NewTerminal(logOutput)
	logger, err := newLogger(terminal)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	flags := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	flags.SetOutput(logOutput)
	flags.Usage = func() {
		fmt.Fprintln(logOutput, "Usage:")
		fmt.Fprintln(logOutput, "  benchmark --scenario PATH [--agent NAME] [--parallel N] [--repeat N]")
		fmt.Fprintln(logOutput, "  benchmark --validate PATH [--parallel N] [--repeat N]")
		fmt.Fprintln(logOutput, "\nOptions:")
		flags.PrintDefaults()
	}
	var scenarioPaths stringList
	flags.Var(&scenarioPaths, "scenario", "path to a scenario YAML file or directory; may be repeated")
	var validationPaths stringList
	flags.Var(&validationPaths, "validate", "path to a validation YAML file or directory; may be repeated")
	agentName := flags.String("agent", string(commandagent.All), "benchmark agent: all, baseline, or prompt")
	parallel := flags.Int("parallel", 1, "maximum number of tasks running at once")
	repeat := flags.Int("repeat", 1, "number of times to run each scenario or validation case")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if len(scenarioPaths) > 0 && len(validationPaths) > 0 {
		return errors.New("scenario and validate paths cannot be combined")
	}
	if len(scenarioPaths) == 0 && len(validationPaths) == 0 {
		return errors.New("scenario or validation path is required; use --scenario PATH or --validate PATH")
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *parallel < 1 {
		return errors.New("parallel must be at least 1")
	}
	if *repeat < 1 {
		return errors.New("repeat must be at least 1")
	}
	agentNames, err := commandagent.Select(*agentName)
	if err != nil {
		return err
	}

	if len(validationPaths) > 0 {
		if strings.ToLower(strings.TrimSpace(*agentName)) != string(commandagent.All) {
			return errors.New("agent selection is only supported with scenario runs")
		}
		return runValidation(ctx, validationPaths, *parallel, *repeat, terminal)
	}
	return runBenchmark(ctx, scenarioPaths, *parallel, *repeat, agentNames, terminal)
}

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("path must not be blank")
	}
	*s = append(*s, value)
	return nil
}

func newLogger(output io.Writer) (*slog.Logger, error) {
	level, err := logging.ParseLevel(os.Getenv("BENCHMARK_LOG_LEVEL"))
	if err != nil {
		return nil, err
	}

	format := logging.Format(os.Getenv("BENCHMARK_LOG_FORMAT"))
	if format == "" {
		format = logging.FormatText
	}
	if color, configured := logColorSetting(); configured {
		return logging.NewWithColor(output, format, level, color)
	}
	if colorProvider, ok := output.(interface{ ColorEnabled() bool }); ok {
		return logging.NewWithColor(output, format, level, colorProvider.ColorEnabled())
	}
	return logging.New(output, format, level)
}

func logColorSetting() (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BENCHMARK_LOG_COLOR"))) {
	case "always", "true", "1":
		return true, true
	case "never", "false", "0":
		return false, true
	}
	return false, os.Getenv("NO_COLOR") != ""
}
