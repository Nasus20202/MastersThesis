package screens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
)

func TestHistogramSpansDataRange(t *testing.T) {
	labels, counts := histogram([]float64{10, 20, 30, 40, 50}, wholeNumber)
	require.Len(t, labels, 5)
	require.Len(t, counts, 5)
	assert.Equal(t, 5, sum(counts))
	assert.Equal(t, "10–18", labels[0])
	assert.Equal(t, "42–50", labels[4])
}

func TestHistogramHandlesSingleValue(t *testing.T) {
	labels, counts := histogram([]float64{7}, wholeNumber)
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

func TestVitalsIncludeDecodingThroughput(t *testing.T) {
	view := &View{}
	metrics := model.Metrics{PredictedTokens: 100, PredictedSeconds: 4, DraftTokens: 50, DraftAccepted: 25}
	out := strings.Join(view.vitals(metrics), " ")
	assert.Contains(t, out, "out t/s")
	assert.Contains(t, out, "25.0")
	assert.Contains(t, out, "draft accept %")
	assert.Contains(t, out, "50%")
}

func TestAgentRankGraphsCombineScoreAndRatePerLine(t *testing.T) {
	metrics := map[string]model.Metrics{
		"skill":    {Attempts: 2, Full: 2, MeanScore: 1},
		"baseline": {Attempts: 2, Full: 1, MeanScore: 0.5},
	}
	out := agentRankGraphs(metrics, 80)
	assert.Contains(t, out, "Mean score · full success by agent")
	lines := strings.Split(out, "\n")
	require.Len(t, lines, 3)
	assert.Contains(t, lines[1], "skill")
	assert.Contains(t, lines[1], "100%")
	assert.Contains(t, lines[2], "baseline")
	assert.Contains(t, lines[2], "50%")
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
