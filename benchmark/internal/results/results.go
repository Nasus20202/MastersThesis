package results

import (
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
)

type RunMetadata struct {
	RunID              string     `json:"run_id"`
	RunType            string     `json:"run_type,omitempty"`
	State              RunState   `json:"state"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	RepositoryRevision string     `json:"repository_revision,omitempty"`
	WorkingTreeDirty   bool       `json:"working_tree_dirty"`
	ExpectedAttempts   int        `json:"expected_attempts"`
	Agents             []string   `json:"agents,omitempty"`
	Parallelism        int        `json:"parallelism"`
	RepeatCount        int        `json:"repeat_count"`
	Scenarios          []string   `json:"scenarios"`
}

type RunState string

const (
	RunStateRunning    RunState = "running"
	RunStateCompleted  RunState = "completed"
	RunStateIncomplete RunState = "incomplete"
)

type RunSummary struct {
	RunID            string                      `json:"run_id"`
	State            RunState                    `json:"state"`
	UpdatedAt        time.Time                   `json:"updated_at"`
	CompletedAt      *time.Time                  `json:"completed_at,omitempty"`
	ExpectedAttempts int                         `json:"expected_attempts"`
	AttemptsRecorded int                         `json:"attempts_recorded"`
	ErrorCount       int                         `json:"error_count"`
	ByCondition      map[string]ConditionSummary `json:"by_condition,omitempty"`
}

type ConditionSummary struct {
	ExpectedAttempts      int                        `json:"expected_attempts"`
	AttemptCount          int                        `json:"attempt_count"`
	FullSuccessCount      int                        `json:"full_success_count"`
	FullSuccessRate       float64                    `json:"full_success_rate"`
	MeanScore             float64                    `json:"mean_score"`
	MacroAverageScore     *float64                   `json:"macro_average_score,omitempty"`
	ErrorCount            int                        `json:"error_count"`
	ValidationPassedCount int                        `json:"validation_passed_count,omitempty"`
	ValidationPassRate    *float64                   `json:"validation_pass_rate,omitempty"`
	Scenarios             map[string]ScenarioSummary `json:"scenarios,omitempty"`
}

type ScenarioSummary struct {
	ExpectedAttempts int     `json:"expected_attempts"`
	AttemptCount     int     `json:"attempt_count"`
	FullSuccessCount int     `json:"full_success_count"`
	FullSuccessRate  float64 `json:"full_success_rate"`
	MeanScore        float64 `json:"mean_score"`
}

type AttemptResult struct {
	RunID      string                         `json:"run_id"`
	Condition  string                         `json:"condition"`
	ScenarioID string                         `json:"scenario_id"`
	Attempt    int                            `json:"attempt"`
	Agent      *common.Result                 `json:"agent,omitempty"`
	Grading    orchestration.GradingResult    `json:"grading"`
	Failure    *orchestration.FailureEvidence `json:"failure,omitempty"`
	Error      string                         `json:"error,omitempty"`
}

type ValidationAttemptResult struct {
	RunID               string                         `json:"run_id"`
	Condition           string                         `json:"condition"`
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
