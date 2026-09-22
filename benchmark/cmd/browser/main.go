// Package main is the benchmark results browser: a read-only terminal UI over
// the runs persisted by the benchmark runner.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/app"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "browser:", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("browser", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage:")
		fmt.Fprintln(output, "  browser [--results PATH] [--scenarios PATH] [--no-watch]")
		fmt.Fprintln(output, "\nOptions:")
		flags.PrintDefaults()
	}
	resultsRoot := flags.String("results", "results", "results directory to browse")
	scenariosRoot := flags.String("scenarios", "scenarios", "scenario corpus used for task titles")
	noWatch := flags.Bool("no-watch", false, "do not watch for run changes")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	root, err := filepath.Abs(*resultsRoot)
	if err != nil {
		return err
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return fmt.Errorf("results directory %q not found", *resultsRoot)
	}
	scenarios, err := filepath.Abs(*scenariosRoot)
	if err != nil {
		return err
	}

	model := app.New(app.Config{ResultsRoot: root, ScenariosRoot: scenarios})
	program := tea.NewProgram(model)

	if !*noWatch {
		messages := make(chan tea.Msg, 8)
		watcher, err := app.Watch(root, func(message tea.Msg) {
			select {
			case messages <- message:
			default:
			}
		})
		if err != nil {
			return err
		}
		defer watcher.Close()
		go func() {
			for message := range messages {
				program.Send(message)
			}
		}()
	}

	_, err = program.Run()
	return err
}
