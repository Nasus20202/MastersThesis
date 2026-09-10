package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/cluster/kind"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/lifecycle"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/logging"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		slog.Error("benchmark failed", "error", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	logger, err := newLogger(output)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	flags := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	flags.SetOutput(output)
	scenarioPath := flags.String("scenario", "", "path to the scenario YAML file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *scenarioPath == "" {
		return errors.New("scenario path is required; use --scenario PATH")
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	definition, err := scenario.Load(*scenarioPath)
	if err != nil {
		return err
	}
	logger.Info("benchmark runner started", "scenario", *scenarioPath)

	executor := command.LocalExecutor{}
	runner := lifecycle.Runner{
		Executor: executor,
		ClusterFactory: func(name string) (lifecycle.Cluster, error) {
			return kind.New(executor, kind.Config{
				Name:       name,
				ConfigPath: definition.Cluster.Kind.ConfigPath(),
			})
		},
	}
	_, err = runner.Run(context.Background(), definition)
	return err
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
	return logging.New(output, format, level)
}
