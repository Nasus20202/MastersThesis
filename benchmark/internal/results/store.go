package results

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
)

type Store struct {
	mu       sync.Mutex
	runDir   string
	metadata RunMetadata
	summary  *summaryAccumulator
	recorded map[string]struct{}
}

func New(root string, metadata RunMetadata) (*Store, error) {
	if strings.TrimSpace(metadata.RunID) == "" {
		return nil, errors.New("result run ID is required")
	}
	if metadata.Parallelism < 1 {
		return nil, errors.New("result parallelism must be at least 1")
	}
	if metadata.RepeatCount < 1 {
		return nil, errors.New("result repeat count must be at least 1")
	}
	if len(metadata.Scenarios) == 0 {
		return nil, errors.New("result requires at least one scenario")
	}
	metadata.Agents = slices.Clone(metadata.Agents)
	metadata.Scenarios = slices.Clone(metadata.Scenarios)
	if metadata.ExpectedAttempts < 1 {
		metadata.ExpectedAttempts = expectedAttemptCount(metadata)
	}
	metadata.RunType = runType(metadata)
	metadata.State = RunStateRunning
	metadata.CompletedAt = nil
	runDir := filepath.Join(root, metadata.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, fmt.Errorf("create result directory: %w", err)
	}
	store := &Store{
		runDir:   runDir,
		metadata: metadata,
		summary:  newSummaryAccumulator(metadata),
		recorded: make(map[string]struct{}),
	}
	if err := store.writeRunMetadata(); err != nil {
		return nil, err
	}
	if err := store.writeSummary(); err != nil {
		return nil, err
	}
	return store, nil
}

// Resume reopens an existing incomplete run and rebuilds its in-memory summary
// from the attempt artifacts already written to disk. Completed attempts are
// retained and can be skipped by the runner; only missing task keys need to be
// scheduled again.
func Resume(root, runID string) (*Store, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, errors.New("result run ID is required")
	}
	runDir := filepath.Join(root, runID)
	metadataData, err := os.ReadFile(filepath.Join(runDir, "run.json"))
	if err != nil {
		return nil, fmt.Errorf("read run metadata: %w", err)
	}
	var metadata RunMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return nil, fmt.Errorf("decode run metadata: %w", err)
	}
	if metadata.RunID != runID {
		return nil, fmt.Errorf("run metadata ID %q does not match %q", metadata.RunID, runID)
	}
	if metadata.State == RunStateCompleted {
		return nil, fmt.Errorf("run %q is already completed", runID)
	}
	if metadata.Parallelism < 1 || metadata.RepeatCount < 1 || len(metadata.Scenarios) == 0 {
		return nil, errors.New("run metadata is incomplete or invalid")
	}
	metadata.Agents = slices.Clone(metadata.Agents)
	metadata.Scenarios = slices.Clone(metadata.Scenarios)
	metadata.RunType = runType(metadata)
	metadata.State = RunStateRunning
	metadata.CompletedAt = nil
	store := &Store{
		runDir:   runDir,
		metadata: metadata,
		summary:  newSummaryAccumulator(metadata),
		recorded: make(map[string]struct{}),
	}
	attemptPaths, err := filepath.Glob(filepath.Join(runDir, "*", "*", "*.json"))
	if err != nil {
		return nil, fmt.Errorf("find run attempts: %w", err)
	}
	for _, path := range attemptPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read run attempt %q: %w", path, err)
		}
		if metadata.RunType == "validation" {
			var artifact ValidationAttemptResult
			if err := json.Unmarshal(data, &artifact); err != nil {
				return nil, fmt.Errorf("decode run attempt %q: %w", path, err)
			}
			if artifact.RunID != metadata.RunID {
				return nil, fmt.Errorf("run attempt %q belongs to %q", path, artifact.RunID)
			}
			store.recorded[attemptKey(artifact.ScenarioID, artifact.CaseID, artifact.Attempt)] = struct{}{}
			passed := artifact.Passed
			store.summary.add(artifact.Condition, artifact.ScenarioID, artifact.Grading.FullSuccess, artifact.Grading.Score, artifact.Error, &passed)
			continue
		}
		var artifact AttemptResult
		if err := json.Unmarshal(data, &artifact); err != nil {
			return nil, fmt.Errorf("decode run attempt %q: %w", path, err)
		}
		if artifact.RunID != metadata.RunID {
			return nil, fmt.Errorf("run attempt %q belongs to %q", path, artifact.RunID)
		}
		store.recorded[attemptKey(artifact.ScenarioID, artifact.Condition, artifact.Attempt)] = struct{}{}
		store.summary.add(artifact.Condition, artifact.ScenarioID, artifact.Grading.FullSuccess, artifact.Grading.Score, artifact.Error, nil)
		store.summary.addThroughput(artifact.Condition, artifact.Agent)
	}
	if err := store.writeRunMetadata(); err != nil {
		return nil, err
	}
	if err := store.writeSummary(); err != nil {
		return nil, err
	}
	return store, nil
}

// HasAttempt reports whether an attempt artifact is already persisted.
func (s *Store) HasAttempt(attempt int, scenarioID, condition string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.recorded[attemptKey(scenarioID, condition, attempt)]
	return exists
}

// Metadata returns the immutable run configuration used to create this store.
func (s *Store) Metadata() RunMetadata {
	s.mu.Lock()
	defer s.mu.Unlock()
	metadata := s.metadata
	metadata.Agents = slices.Clone(metadata.Agents)
	metadata.Scenarios = slices.Clone(metadata.Scenarios)
	return metadata
}

func attemptKey(scenarioID, condition string, attempt int) string {
	return scenarioID + "\x00" + condition + "\x00" + strconv.Itoa(attempt)
}

func (s *Store) WriteAttempt(attempt int, agent string, result orchestration.RunResult) error {
	return s.writeAttempt(attempt, result.ScenarioID, agent, result, nil)
}

func (s *Store) WriteAttemptFailure(attempt int, scenarioID, agent string, result orchestration.RunResult, runErr error) error {
	if runErr == nil {
		return errors.New("result execution error is required")
	}
	return s.writeAttempt(attempt, scenarioID, agent, result, runErr)
}

func (s *Store) writeAttempt(attempt int, scenarioID, agent string, result orchestration.RunResult, runErr error) error {
	if attempt < 1 {
		return errors.New("result attempt number must be at least 1")
	}
	if strings.TrimSpace(scenarioID) == "" {
		return errors.New("result scenario ID is required")
	}
	if strings.TrimSpace(agent) == "" {
		return errors.New("result agent is required")
	}

	scenarioDir := filepath.Join(s.runDir, scenarioID, agent)
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		return fmt.Errorf("create scenario result directory: %w", err)
	}
	path := filepath.Join(scenarioDir, fmt.Sprintf("%03d.json", attempt))
	condition := result.Condition
	if strings.TrimSpace(condition) == "" {
		condition = agent
	}
	artifact := AttemptResult{
		RunID:      s.metadata.RunID,
		Condition:  condition,
		ScenarioID: scenarioID,
		Attempt:    attempt,
		Agent:      result.Agent,
		Grading:    result.Grading,
		Failure:    result.Failure,
	}
	if runErr != nil {
		artifact.Error = runErr.Error()
	}
	return s.recordAttempt(path, artifact, condition, condition, scenarioID, result.Grading.FullSuccess, result.Grading.Score, artifact.Error, nil)
}

func (s *Store) WriteValidationAttempt(attempt int, scenarioID, caseID string, expectedScore float64, expectedFullSuccess bool, result orchestration.RunResult, validationErr error) error {
	if attempt < 1 {
		return errors.New("result attempt number must be at least 1")
	}
	if strings.TrimSpace(scenarioID) == "" {
		return errors.New("result scenario ID is required")
	}
	if strings.TrimSpace(caseID) == "" {
		return errors.New("result case ID is required")
	}

	caseDir := filepath.Join(s.runDir, scenarioID, caseID)
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		return fmt.Errorf("create validation result directory: %w", err)
	}
	condition := result.Condition
	if strings.TrimSpace(condition) == "" {
		condition = "validation"
	}
	artifact := ValidationAttemptResult{
		RunID:               s.metadata.RunID,
		Condition:           condition,
		ScenarioID:          scenarioID,
		CaseID:              caseID,
		Attempt:             attempt,
		ExpectedScore:       expectedScore,
		ExpectedFullSuccess: expectedFullSuccess,
		Grading:             result.Grading,
		Failure:             result.Failure,
		Passed:              validationErr == nil,
	}
	if validationErr != nil {
		artifact.Error = validationErr.Error()
	}
	passed := artifact.Passed
	path := filepath.Join(caseDir, fmt.Sprintf("%03d.json", attempt))
	return s.recordAttempt(path, artifact, condition, caseID, scenarioID, result.Grading.FullSuccess, result.Grading.Score, artifact.Error, &passed)
}

func (s *Store) recordAttempt(path string, artifact any, condition, recordKey, scenarioID string, fullSuccess bool, score float64, errorText string, validationPassed *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metadata.State != RunStateRunning {
		return errors.New("cannot write attempt result after run finalization")
	}
	if _, exists := s.recorded[attemptKey(scenarioID, recordKey, artifactAttempt(artifact))]; exists {
		return fmt.Errorf("attempt result already recorded: %s", path)
	}
	if err := writeJSON(path, artifact); err != nil {
		return fmt.Errorf("write attempt result: %w", err)
	}
	s.recorded[attemptKey(scenarioID, recordKey, artifactAttempt(artifact))] = struct{}{}
	s.summary.add(condition, scenarioID, fullSuccess, score, errorText, validationPassed)
	s.summary.addThroughput(condition, artifactAgent(artifact))
	if err := s.writeSummary(); err != nil {
		return err
	}
	return nil
}

func artifactAttempt(artifact any) int {
	switch value := artifact.(type) {
	case AttemptResult:
		return value.Attempt
	case ValidationAttemptResult:
		return value.Attempt
	default:
		return 0
	}
}

// artifactAgent returns the model-loop evidence of an attempt, or nil for
// validation attempts, which run no model.
func artifactAgent(artifact any) *common.Result {
	if value, ok := artifact.(AttemptResult); ok {
		return value.Agent
	}
	return nil
}

func (s *Store) Finalize(completedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	completedAt = completedAt.UTC()
	s.metadata.CompletedAt = &completedAt
	if s.summary.attemptsRecorded >= s.summary.metadata.ExpectedAttempts {
		s.metadata.State = RunStateCompleted
	} else {
		s.metadata.State = RunStateIncomplete
	}
	if err := s.writeSummary(); err != nil {
		return err
	}
	return s.writeRunMetadata()
}

func (s *Store) writeRunMetadata() error {
	if err := writeJSON(filepath.Join(s.runDir, "run.json"), s.metadata); err != nil {
		return fmt.Errorf("write run metadata: %w", err)
	}
	return nil
}

func (s *Store) writeSummary() error {
	summary := s.summary.summary(s.metadata.State, s.metadata.CompletedAt)
	if err := writeJSON(filepath.Join(s.runDir, "results.json"), summary); err != nil {
		return fmt.Errorf("write run results: %w", err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return nil
}
