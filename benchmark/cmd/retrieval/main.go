// Package main is the retrieval CLI for building, evaluating and searching the
// RAG indexes.
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
)

const usage = `Usage:
  retrieval build-index [--config PATH ...] [--chunking sections|windows|all]
  retrieval generate-queries --probes PATH --out PATH
  retrieval evaluate [--config PATH ...] --queries PATH
  retrieval search [--config PATH ...] [--mode MODE] [--chunking CHUNKING] [--k N] QUERY`

func main() {
	os.Exit(runMain())
}

func runMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "retrieval:", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, args []string, output, logOutput io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(logOutput, usage)
		return flag.ErrHelp
	}
	name, args := args[0], args[1:]
	flags := flag.NewFlagSet("retrieval "+name, flag.ContinueOnError)
	flags.SetOutput(logOutput)
	flags.Usage = func() {
		fmt.Fprintln(logOutput, usage)
		fmt.Fprintln(logOutput, "\nOptions:")
		flags.PrintDefaults()
	}
	command, ok := commands[name]
	if !ok {
		fmt.Fprintln(logOutput, usage)
		return fmt.Errorf("unknown command %q", name)
	}
	return command(ctx, flags, args, output)
}

type command func(context.Context, *flag.FlagSet, []string, io.Writer) error

var commands = map[string]command{
	"build-index":      runBuildIndex,
	"generate-queries": runGenerateQueries,
	"evaluate":         runEvaluate,
	"search":           runSearch,
}

func parse(flags *flag.FlagSet, args []string) error {
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	return nil
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func configFlag(flags *flag.FlagSet) *stringList {
	var paths stringList
	flags.Var(&paths, "config", "load benchmark YAML configuration; may be repeated in overlay order")
	return &paths
}
