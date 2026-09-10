package lifecycle

import (
	"strings"
	"testing"
)

func TestCalculateGradingResult(t *testing.T) {
	tests := []struct {
		name        string
		criteria    []CriterionResult
		wantScore   float64
		wantSuccess bool
	}{
		{
			name: "all pass",
			criteria: []CriterionResult{
				{ID: "first", Weight: 1, Passed: true},
				{ID: "second", Weight: 3, Passed: true},
			},
			wantScore:   1,
			wantSuccess: true,
		},
		{
			name: "all fail",
			criteria: []CriterionResult{
				{ID: "first", Weight: 1},
				{ID: "second", Weight: 3},
			},
			wantScore:   0,
			wantSuccess: false,
		},
		{
			name: "partial success uses weights",
			criteria: []CriterionResult{
				{ID: "first", Weight: 1, Passed: true},
				{ID: "second", Weight: 3},
			},
			wantScore:   0.25,
			wantSuccess: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := calculateGradingResult(test.criteria)
			if err != nil {
				t.Fatalf("calculate grading result: %v", err)
			}
			if result.Score != test.wantScore {
				t.Fatalf("score = %v, want %v", result.Score, test.wantScore)
			}
			if result.FullSuccess != test.wantSuccess {
				t.Fatalf("full success = %t, want %t", result.FullSuccess, test.wantSuccess)
			}
		})
	}
}

func TestCalculateGradingResultPreservesEvidence(t *testing.T) {
	criteria := []CriterionResult{{
		ID:              "workload-restored",
		Weight:          2,
		Stdout:          "observed stdout",
		Stderr:          "observed stderr",
		ExitCode:        7,
		DurationSeconds: 0.15,
	}}

	result, err := calculateGradingResult(criteria)
	if err != nil {
		t.Fatalf("calculate grading result: %v", err)
	}
	if len(result.Criteria) != 1 {
		t.Fatalf("criteria = %#v, want one result", result.Criteria)
	}
	got := result.Criteria[0]
	if got.ID != criteria[0].ID || got.Weight != criteria[0].Weight || got.Passed != criteria[0].Passed {
		t.Fatalf("criterion metadata = %#v, want %#v", got, criteria[0])
	}
	if got.Stdout != criteria[0].Stdout || got.Stderr != criteria[0].Stderr {
		t.Fatalf("criterion output = %#v, want %#v", got, criteria[0])
	}
	if got.ExitCode != criteria[0].ExitCode || got.DurationSeconds != criteria[0].DurationSeconds {
		t.Fatalf("criterion execution data = %#v, want %#v", got, criteria[0])
	}
}

func TestCalculateGradingResultRejectsInvalidInput(t *testing.T) {
	if _, err := calculateGradingResult(nil); err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("empty criteria error = %v, want missing criteria error", err)
	}

	if _, err := calculateGradingResult([]CriterionResult{{ID: "invalid", Weight: 0}}); err == nil || !strings.Contains(err.Error(), "non-positive") {
		t.Fatalf("zero weight error = %v, want non-positive weight error", err)
	}
}
