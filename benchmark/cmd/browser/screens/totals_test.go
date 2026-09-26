package screens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func TestTotalsAggregateEveryRun(t *testing.T) {
	root := t.TempDir()
	for _, runID := range []string{"run-1", "run-2"} {
		store, err := results.New(root, results.RunMetadata{RunID: runID, Agents: []string{"skill"}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
		require.NoError(t, err)
		result := orchestration.RunResult{
			ScenarioID: "alpha", Condition: "skill",
			Agent: &common.Result{
				Termination:     common.TerminationCompleted,
				DurationSeconds: 10,
				TokenUsage:      common.TokenUsage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150, CachedTokens: 25},
			},
			Grading: orchestration.GradingResult{Score: 1, FullSuccess: true},
		}
		require.NoError(t, store.WriteAttempt(1, "skill", result))
	}

	read := model.NewStore(model.StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())
	view := NewView(read)

	out := strings.Join(view.buildTotals(120), "\n")
	assert.Contains(t, out, "All runs")
	assert.Contains(t, out, "attempts")
	assert.Contains(t, out, "300")
	assert.Contains(t, out, "Tokens")
	assert.Contains(t, out, "Termination")
}
