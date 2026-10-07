package results

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"time"
)

// Merge combines runs of one configuration, such as the shards of a benchmark
// or repeated runs of the same scenarios, into a new run runID under root.
// Attempts are copied with the new run ID, numbered after the attempts earlier
// runs gave the same scenario, and the summary is rebuilt from them; the
// sources are left unchanged.
func Merge(root, runID string, sourceIDs []string) (RunMetadata, error) {
	if len(sourceIDs) == 0 {
		return RunMetadata{}, errors.New("merge requires at least one run")
	}
	sources := make([]RunMetadata, len(sourceIDs))
	for i, sourceID := range sourceIDs {
		if slices.Contains(sourceIDs[:i], sourceID) {
			return RunMetadata{}, fmt.Errorf("run %q is given more than once", sourceID)
		}
		metadata, err := readRunMetadata(filepath.Join(root, sourceID))
		if err != nil {
			return RunMetadata{}, fmt.Errorf("read run %q: %w", sourceID, err)
		}
		if i > 0 && !reflect.DeepEqual(configuration(metadata), configuration(sources[0])) {
			return RunMetadata{}, fmt.Errorf("run %q differs from %q in its configuration", sourceID, sourceIDs[0])
		}
		sources[i] = metadata
	}

	merged := sources[0]
	merged.RunID = runID
	merged.Scenarios, merged.WorkingTreeDirty = nil, false
	merged.Parallelism = peakParallelism(sources)
	completedAt := sources[0].StartedAt
	// offsets[i][scenario] is the number of repeats runs before source i gave
	// the scenario; its attempts are numbered after them.
	repeats := make(map[string]int)
	offsets := make([]map[string]int, len(sources))
	for i, source := range sources {
		offsets[i] = make(map[string]int, len(source.Scenarios))
		for _, scenarioID := range source.Scenarios {
			if _, seen := repeats[scenarioID]; !seen {
				merged.Scenarios = append(merged.Scenarios, scenarioID)
			}
			offsets[i][scenarioID] = repeats[scenarioID]
			repeats[scenarioID] += source.RepeatCount
		}
		merged.WorkingTreeDirty = merged.WorkingTreeDirty || source.WorkingTreeDirty
		if source.StartedAt.Before(merged.StartedAt) {
			merged.StartedAt = source.StartedAt
		}
		if source.CompletedAt != nil && source.CompletedAt.After(completedAt) {
			completedAt = *source.CompletedAt
		}
	}
	// Scenarios with fewer repeats than the most repeated one leave the run
	// incomplete and without a macro average.
	merged.RepeatCount = 0
	for _, count := range repeats {
		merged.RepeatCount = max(merged.RepeatCount, count)
	}
	merged.ExpectedAttempts = 0
	merged.ExpectedAttempts = expectedAttemptCount(merged)

	runDir := filepath.Join(root, runID)
	if err := os.Mkdir(runDir, 0o755); err != nil {
		return RunMetadata{}, fmt.Errorf("create merged run directory: %w", err)
	}
	store := newStore(runDir, merged)
	for i, sourceID := range sourceIDs {
		if err := store.copyAttempts(filepath.Join(root, sourceID), offsets[i]); err != nil {
			return RunMetadata{}, err
		}
	}
	// A run is completed only once every expected attempt is recorded, so a
	// merge of an unfinished shard is incomplete.
	if err := store.Finalize(completedAt); err != nil {
		return RunMetadata{}, err
	}
	return store.Metadata(), nil
}

// peakParallelism is the most tasks the runs allowed at once: the parallelism
// of the runs in progress when one of them started, such as concurrent shards.
func peakParallelism(sources []RunMetadata) int {
	peak := 0
	for _, start := range sources {
		total := 0
		for _, source := range sources {
			if !source.StartedAt.After(start.StartedAt) && (source.CompletedAt == nil || !source.CompletedAt.Before(start.StartedAt)) {
				total += source.Parallelism
			}
		}
		peak = max(peak, total)
	}
	return peak
}

// configuration is the run metadata that must match for runs to be merged.
func configuration(metadata RunMetadata) RunMetadata {
	metadata.RunID, metadata.State = "", ""
	metadata.StartedAt, metadata.CompletedAt = time.Time{}, nil
	metadata.WorkingTreeDirty, metadata.ExpectedAttempts, metadata.Parallelism = false, 0, 0
	metadata.Scenarios, metadata.RepeatCount = nil, 0
	metadata.RunType = runType(metadata)
	return metadata
}

func (s *Store) copyAttempts(sourceDir string, offsets map[string]int) error {
	refs, err := listAttempts(sourceDir)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		attempt, err := LoadAttempt(ref, s.metadata.RunType)
		if err != nil {
			return err
		}
		number := offsets[ref.ScenarioID] + ref.Attempt
		if attempt.Benchmark != nil {
			attempt.Benchmark.RunID, attempt.Benchmark.Attempt = s.metadata.RunID, number
		}
		if attempt.Validation != nil {
			attempt.Validation.RunID, attempt.Validation.Attempt = s.metadata.RunID, number
		}
		path, err := s.attemptPath(number, ref.ScenarioID, ref.Group, "merged")
		if err != nil {
			return err
		}
		if err := s.recordAttempt(path, attempt); err != nil {
			return err
		}
	}
	return nil
}
