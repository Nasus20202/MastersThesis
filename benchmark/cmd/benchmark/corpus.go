package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/scenario"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/validation"
)

// runCorpusCheck loads every scenario in the corpus and its validation cases so
// a broken scenario, an ID/directory mismatch or a missing validation file
// fails the check instead of only failing at run time.
func runCorpusCheck(root string) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve corpus path %q: %w", root, err)
	}

	scenarioPaths, err := scenario.Discover(absoluteRoot)
	if err != nil {
		return err
	}
	validationPaths, err := validation.Discover(absoluteRoot)
	if err != nil {
		return err
	}
	if len(scenarioPaths) == 0 {
		return fmt.Errorf("no scenarios found under %q", root)
	}

	// A scenario directory holds one scenario definition and one validation
	// file. Comparing the two sets by directory reports a missing or
	// unexpectedly named file instead of silently skipping it.
	scenariosByDir := indexByDir(scenarioPaths)
	validationsByDir := indexByDir(validationPaths)
	var errs []error
	for _, dir := range sortedKeys(scenariosByDir) {
		if _, ok := validationsByDir[dir]; !ok {
			errs = append(errs, fmt.Errorf("%s: missing validation.yaml", dir))
		}
	}
	for _, dir := range sortedKeys(validationsByDir) {
		if _, ok := scenariosByDir[dir]; !ok {
			errs = append(errs, fmt.Errorf("%s: missing scenario.yaml", dir))
		}
	}

	// Loading each scenario logs at info level; keep the check output to a
	// single summary line unless something fails.
	logger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer slog.SetDefault(logger)

	for _, dir := range sortedKeys(scenariosByDir) {
		definition, err := scenario.Load(scenariosByDir[dir])
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		if filepath.Base(dir) != definition.ID {
			errs = append(errs, fmt.Errorf("%s: scenario id %q does not match directory name", dir, definition.ID))
		}
		cases, err := validation.LoadCases([]string{validationsByDir[dir]})
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		if len(cases) == 0 {
			errs = append(errs, fmt.Errorf("%s: validation has no cases", dir))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	logger.Info("scenario corpus valid", "root", root, "scenarios", len(scenariosByDir))
	return nil
}

func indexByDir(paths []string) map[string]string {
	index := make(map[string]string, len(paths))
	for _, path := range paths {
		index[filepath.Dir(path)] = path
	}
	return index
}

func sortedKeys(index map[string]string) []string {
	keys := make([]string, 0, len(index))
	for key := range index {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
