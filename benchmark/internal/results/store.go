package results

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
)

type Store struct {
	runDir   string
	metadata RunMetadata
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
	metadata.Scenarios = slices.Clone(metadata.Scenarios)
	runDir := filepath.Join(root, metadata.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, fmt.Errorf("create result directory: %w", err)
	}
	store := &Store{runDir: runDir, metadata: metadata}
	if err := store.writeRunMetadata(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) WriteAttempt(attempt int, result orchestration.RunResult) error {
	return s.writeAttempt(attempt, result.ScenarioID, result, nil)
}

func (s *Store) WriteAttemptFailure(attempt int, scenarioID string, result orchestration.RunResult, runErr error) error {
	if runErr == nil {
		return errors.New("result execution error is required")
	}
	return s.writeAttempt(attempt, scenarioID, result, runErr)
}

func (s *Store) writeAttempt(attempt int, scenarioID string, result orchestration.RunResult, runErr error) error {
	if attempt < 1 {
		return errors.New("result attempt number must be at least 1")
	}
	if strings.TrimSpace(scenarioID) == "" {
		return errors.New("result scenario ID is required")
	}

	scenarioDir := filepath.Join(s.runDir, scenarioID)
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		return fmt.Errorf("create scenario result directory: %w", err)
	}
	path := filepath.Join(scenarioDir, fmt.Sprintf("%03d.json", attempt))
	artifact := AttemptResult{
		RunID:      s.metadata.RunID,
		ScenarioID: scenarioID,
		Attempt:    attempt,
		Grading:    result.Grading,
		Failure:    result.Failure,
	}
	if runErr != nil {
		artifact.Error = runErr.Error()
	}
	if err := writeJSON(path, artifact); err != nil {
		return fmt.Errorf("write attempt result: %w", err)
	}
	return nil
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
	artifact := ValidationAttemptResult{
		RunID:               s.metadata.RunID,
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
	if err := writeJSON(filepath.Join(caseDir, fmt.Sprintf("%03d.json", attempt)), artifact); err != nil {
		return fmt.Errorf("write validation result: %w", err)
	}
	return nil
}

func (s *Store) Finalize(completedAt time.Time) error {
	s.metadata.CompletedAt = &completedAt
	return s.writeRunMetadata()
}

func (s *Store) writeRunMetadata() error {
	if err := writeJSON(filepath.Join(s.runDir, "run.json"), s.metadata); err != nil {
		return fmt.Errorf("write run metadata: %w", err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
