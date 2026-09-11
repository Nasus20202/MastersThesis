package results

import (
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
)

type RunMetadata struct {
	RunID       string     `json:"run_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Parallelism int        `json:"parallelism"`
	RepeatCount int        `json:"repeat_count"`
	Scenarios   []string   `json:"scenarios"`
}

type AttemptResult struct {
	RunID      string                         `json:"run_id"`
	ScenarioID string                         `json:"scenario_id"`
	Attempt    int                            `json:"attempt"`
	Agent      *common.Result                 `json:"agent,omitempty"`
	Grading    orchestration.GradingResult    `json:"grading"`
	Failure    *orchestration.FailureEvidence `json:"failure,omitempty"`
	Error      string                         `json:"error,omitempty"`
}

type ValidationAttemptResult struct {
	RunID               string                         `json:"run_id"`
	ScenarioID          string                         `json:"scenario_id"`
	CaseID              string                         `json:"case_id"`
	Attempt             int                            `json:"attempt"`
	ExpectedScore       float64                        `json:"expected_score"`
	ExpectedFullSuccess bool                           `json:"expected_full_success"`
	Grading             orchestration.GradingResult    `json:"grading"`
	Failure             *orchestration.FailureEvidence `json:"failure,omitempty"`
	Passed              bool                           `json:"passed"`
	Error               string                         `json:"error,omitempty"`
}
