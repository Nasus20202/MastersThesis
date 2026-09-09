package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/logging"
)

func main() {
	logger, err := newLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	slog.SetDefault(logger)
	slog.Info("benchmark runner started")
}

func newLogger() (*slog.Logger, error) {
	level, err := logging.ParseLevel(os.Getenv("BENCHMARK_LOG_LEVEL"))
	if err != nil {
		return nil, err
	}

	format := logging.Format(os.Getenv("BENCHMARK_LOG_FORMAT"))
	if format == "" {
		format = logging.FormatText
	}
	return logging.New(os.Stderr, format, level)
}
