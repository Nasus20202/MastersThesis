package results

// This file holds read-only views over the on-disk result tree. Nothing here
// writes; the loader only decodes artifacts written by the benchmark runner so
// tools such as the browser can inspect history without mutating it.

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
)

// RunRef is a lightweight reference to a run discovered on disk.
type RunRef struct {
	RunID    string
	Metadata RunMetadata
	Summary  *RunSummary
}

// AttemptRef identifies one persisted attempt without loading its payload. The
// Group is the condition directory for benchmark runs and the case directory
// for validation runs.
type AttemptRef struct {
	ScenarioID string
	Group      string
	Attempt    int
	Path       string
}

// RunSnapshot is a read-only view of a run: metadata, summary and the index of
// attempt artifacts. Attempt payloads are loaded on demand.
type RunSnapshot struct {
	Root     string
	RunID    string
	Metadata RunMetadata
	Summary  *RunSummary
	Attempts []AttemptRef
}

// Attempt is one decoded attempt artifact. Exactly one of Benchmark or
// Validation is set, selected by the run type.
type Attempt struct {
	Ref        AttemptRef
	Benchmark  *AttemptResult
	Validation *ValidationAttemptResult
}

// ListRuns reads every run directory directly under root and returns them
// newest first. Directories without run metadata are ignored.
func ListRuns(root string) ([]RunRef, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read results directory %q: %w", root, err)
	}
	runs := make([]RunRef, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runDir := filepath.Join(root, entry.Name())
		metadata, err := readRunMetadata(runDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		summary, err := readSummary(runDir)
		if err != nil {
			return nil, err
		}
		runs = append(runs, RunRef{RunID: metadata.RunID, Metadata: metadata, Summary: summary})
	}
	slices.SortFunc(runs, func(a, b RunRef) int {
		if !a.Metadata.StartedAt.Equal(b.Metadata.StartedAt) {
			return b.Metadata.StartedAt.Compare(a.Metadata.StartedAt)
		}
		return strings.Compare(b.RunID, a.RunID)
	})
	return runs, nil
}

// LoadRun reads one run's metadata, summary and attempt index. When the
// summary file is missing, as in an interrupted run, it is rebuilt from the
// persisted attempts.
func LoadRun(root, runID string) (RunSnapshot, error) {
	runDir := filepath.Join(root, runID)
	metadata, err := readRunMetadata(runDir)
	if err != nil {
		return RunSnapshot{}, err
	}
	if metadata.RunID != runID {
		return RunSnapshot{}, fmt.Errorf("run metadata ID %q does not match %q", metadata.RunID, runID)
	}
	attempts, err := listAttempts(runDir)
	if err != nil {
		return RunSnapshot{}, err
	}
	summary, err := readSummary(runDir)
	if err != nil {
		return RunSnapshot{}, err
	}
	if summary == nil {
		summary, err = deriveSummary(metadata, attempts)
		if err != nil {
			return RunSnapshot{}, err
		}
	}
	return RunSnapshot{
		Root:     root,
		RunID:    runID,
		Metadata: metadata,
		Summary:  summary,
		Attempts: attempts,
	}, nil
}

// LoadAttempt decodes the attempt referenced by ref using the run's type.
func LoadAttempt(root, runID string, ref AttemptRef, runType string) (Attempt, error) {
	data, err := os.ReadFile(ref.Path)
	if err != nil {
		return Attempt{}, fmt.Errorf("read attempt %q: %w", ref.Path, err)
	}
	attempt := Attempt{Ref: ref}
	if runType == "validation" {
		var artifact ValidationAttemptResult
		if err := json.Unmarshal(data, &artifact); err != nil {
			return Attempt{}, fmt.Errorf("decode attempt %q: %w", ref.Path, err)
		}
		attempt.Validation = &artifact
		return attempt, nil
	}
	var artifact AttemptResult
	if err := json.Unmarshal(data, &artifact); err != nil {
		return Attempt{}, fmt.Errorf("decode attempt %q: %w", ref.Path, err)
	}
	attempt.Benchmark = &artifact
	return attempt, nil
}

// Grading returns the grading outcome of whichever artifact is set.
func (a Attempt) Grading() orchestration.GradingResult {
	if a.Validation != nil {
		return a.Validation.Grading
	}
	if a.Benchmark != nil {
		return a.Benchmark.Grading
	}
	return orchestration.GradingResult{}
}

// Failure returns the lifecycle failure evidence of whichever artifact is set.
func (a Attempt) Failure() *orchestration.FailureEvidence {
	if a.Validation != nil {
		return a.Validation.Failure
	}
	if a.Benchmark != nil {
		return a.Benchmark.Failure
	}
	return nil
}

// Error returns the recorded execution error of whichever artifact is set.
func (a Attempt) Error() string {
	if a.Validation != nil {
		return a.Validation.Error
	}
	if a.Benchmark != nil {
		return a.Benchmark.Error
	}
	return ""
}

// Fingerprint summarizes file sizes and modification times under a run so
// callers can detect changes without reloading everything.
func Fingerprint(root, runID string) (string, error) {
	runDir := filepath.Join(root, runID)
	hash := fnv.New64a()
	err := filepath.WalkDir(runDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(runDir, path)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(hash, "%s\x00%d\x00%d\n", rel, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("fingerprint run %q: %w", runID, err)
	}
	return strconv.FormatUint(hash.Sum64(), 16), nil
}

func readRunMetadata(runDir string) (RunMetadata, error) {
	var metadata RunMetadata
	if err := readJSON(filepath.Join(runDir, "run.json"), &metadata); err != nil {
		return RunMetadata{}, err
	}
	return metadata, nil
}

func readSummary(runDir string) (*RunSummary, error) {
	var summary RunSummary
	if err := readJSON(filepath.Join(runDir, "results.json"), &summary); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return &summary, nil
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("decode %q: %w", path, err)
	}
	return nil
}

func listAttempts(runDir string) ([]AttemptRef, error) {
	paths, err := filepath.Glob(filepath.Join(runDir, "*", "*", "*.json"))
	if err != nil {
		return nil, fmt.Errorf("find run attempts: %w", err)
	}
	attempts := make([]AttemptRef, 0, len(paths))
	for _, path := range paths {
		ref, err := parseAttemptPath(runDir, path)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, ref)
	}
	slices.SortFunc(attempts, func(a, b AttemptRef) int {
		if a.ScenarioID != b.ScenarioID {
			return strings.Compare(a.ScenarioID, b.ScenarioID)
		}
		if a.Group != b.Group {
			return strings.Compare(a.Group, b.Group)
		}
		return a.Attempt - b.Attempt
	})
	return attempts, nil
}

func parseAttemptPath(runDir, path string) (AttemptRef, error) {
	rel, err := filepath.Rel(runDir, path)
	if err != nil {
		return AttemptRef{}, fmt.Errorf("resolve attempt path %q: %w", path, err)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) != 3 {
		return AttemptRef{}, fmt.Errorf("unexpected attempt path %q", rel)
	}
	attempt, err := strconv.Atoi(strings.TrimSuffix(parts[2], ".json"))
	if err != nil {
		return AttemptRef{}, fmt.Errorf("parse attempt number in %q: %w", rel, err)
	}
	return AttemptRef{ScenarioID: parts[0], Group: parts[1], Attempt: attempt, Path: path}, nil
}

func deriveSummary(metadata RunMetadata, attempts []AttemptRef) (*RunSummary, error) {
	accumulator := newSummaryAccumulator(metadata)
	for _, ref := range attempts {
		data, err := os.ReadFile(ref.Path)
		if err != nil {
			return nil, fmt.Errorf("read attempt %q: %w", ref.Path, err)
		}
		if runType(metadata) == "validation" {
			var artifact ValidationAttemptResult
			if err := json.Unmarshal(data, &artifact); err != nil {
				return nil, fmt.Errorf("decode attempt %q: %w", ref.Path, err)
			}
			passed := artifact.Passed
			accumulator.add(artifact.Condition, artifact.ScenarioID, artifact.Grading.FullSuccess, artifact.Grading.Score, artifact.Error, &passed)
			continue
		}
		var artifact AttemptResult
		if err := json.Unmarshal(data, &artifact); err != nil {
			return nil, fmt.Errorf("decode attempt %q: %w", ref.Path, err)
		}
		accumulator.add(artifact.Condition, artifact.ScenarioID, artifact.Grading.FullSuccess, artifact.Grading.Score, artifact.Error, nil)
	}
	state := metadata.State
	if state == "" {
		state = RunStateIncomplete
	}
	summary := accumulator.summary(state, metadata.CompletedAt)
	return &summary, nil
}
