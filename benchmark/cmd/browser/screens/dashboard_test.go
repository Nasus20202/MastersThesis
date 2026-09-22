package screens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
)

func TestHistogramSpansDataRange(t *testing.T) {
	labels, counts := histogram([]float64{10, 20, 30, 40, 50}, 5, wholeNumber)
	require.Len(t, labels, 5)
	require.Len(t, counts, 5)
	assert.Equal(t, 5, sum(counts))
	assert.Equal(t, "10–18", labels[0])
	assert.Equal(t, "42–50", labels[4])
}

func TestHistogramHandlesSingleValue(t *testing.T) {
	labels, counts := histogram([]float64{7}, 5, wholeNumber)
	assert.Equal(t, []string{"7"}, labels)
	assert.Equal(t, []int{1}, counts)
}

func sum(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func TestCriteriaMatrixRendersCells(t *testing.T) {
	matrix := model.CriteriaMatrix{
		Agents: []string{"baseline", "skill"},
		Rows: []model.CriteriaRow{
			{ID: "ready", Cells: map[string]model.CriterionStat{"baseline": {Total: 1}, "skill": {Passed: 1, Total: 1}}},
		},
	}
	out := criteriaMatrix(matrix, 80)
	assert.Contains(t, out, "CRITERION")
	assert.Contains(t, out, "baseline")
	assert.Contains(t, out, "0/1")
	assert.Contains(t, out, "1/1")
}
