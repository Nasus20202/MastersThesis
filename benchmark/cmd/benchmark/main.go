package main

import (
	"context"
	"encoding/json"
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
	"github.com/Nasus20202/MastersThesis/benchmark/internal/sandbox"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
)

const (
	sandboxImage          = "masters-thesis-sandbox:ubuntu-26.04"
	sandboxDockerfilePath = "sandbox/Dockerfile"
	sandboxBuildContext   = "sandbox"
)

func main() {
	if err := run(os.Args[1:], os.Stderr, os.Stdout); err != nil {
		slog.Error("benchmark failed", "error", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, logOutput, resultOutput io.Writer) error {
	logger, err := newLogger(logOutput)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	flags := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	flags.SetOutput(logOutput)
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
		SandboxFactory: func(name, kubeconfigPath string) (lifecycle.Sandbox, error) {
			return sandbox.New(executor, sandbox.Config{
				Name:           name + "-sandbox",
				Image:          sandboxImage,
				DockerfilePath: sandboxDockerfilePath,
				BuildContext:   sandboxBuildContext,
				KubeconfigPath: kubeconfigPath,
			})
		},
	}
	result, err := runner.Run(context.Background(), definition)
	if err != nil {
		return err
	}
	return writeResult(resultOutput, result)
}

func writeResult(output io.Writer, result lifecycle.RunResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("write benchmark result: %w", err)
	}
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
	return logging.New(output, format, level)
}
