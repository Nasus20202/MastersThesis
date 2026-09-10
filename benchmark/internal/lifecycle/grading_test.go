package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
			assert.Equal(t, test.wantScore, result.Score)
			assert.Equal(t, test.wantSuccess, result.FullSuccess)
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
	require.NoError(t, err)
	require.Len(t, result.Criteria, 1)
	assert.Equal(t, criteria[0], result.Criteria[0])
}

func TestCalculateGradingResultRejectsInvalidInput(t *testing.T) {
	_, err := calculateGradingResult(nil)
	assert.ErrorContains(t, err, "at least one")

	_, err = calculateGradingResult([]CriterionResult{{ID: "invalid", Weight: 0}})
	assert.ErrorContains(t, err, "non-positive")
}
